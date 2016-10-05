package command

import (
	"bufio"
	"bytes"
	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
	"github.com/k0kubun/pp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type tomlConfig struct {
	Database map[string]Database
	SSH      map[string]SSH
}

// Database settings
type Database struct {
	Host             string
	Port             int
	ManagementSystem string
	Name             string
	User             string
	Password         string
}

// SSH settings
type SSH struct {
	Host string
	Port string
	User string
	Key  string
}

// CmdSync supports `sync` command in CLI
func CmdSync(c *cli.Context) {
	var tmlconf tomlConfig
	if _, err := toml.DecodeFile(c.String("config"), &tmlconf); err != nil {
		// TODO: pkg/errors
		pp.Print(err)
	}

	fromDBConf := tmlconf.Database[c.String("from")]
	// toDBConf := tmlconf.Database[c.String("to")]
	fromSSHConf := tmlconf.SSH[c.String("from")]
	// toSSHConf := tmlconf.SSH[c.String("to")]

	syncTimestamp := strconv.FormatInt(time.Now().Unix(), 10)

	usr, _ := user.Current()
	keypathString := strings.Replace(fromSSHConf.Key, "~", usr.HomeDir, 1)
	keypath, _ := filepath.Abs(keypathString)
	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: fromSSHConf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	conn, err := ssh.Dial("tcp", fromSSHConf.Host+":"+fromSSHConf.Port, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var listTableStdoutBuf bytes.Buffer
	session.Stdout = &listTableStdoutBuf
	listTableCmd := "mysql " + fromDBConf.Name + " -u" + fromDBConf.User + " -p" + fromDBConf.Password + " -B -N -e 'show tables'"

	err = session.Run(listTableCmd)

	loadDirName := "/tmp/db_sync_" + syncTimestamp
	pp.Print(loadDirName)
	if err := os.MkdirAll(loadDirName, 0777); err != nil {
		pp.Fatal(err)
	}

	listTableResultFile := loadDirName + "/table_list.txt"
	pp.Print(listTableResultFile)
	ioutil.WriteFile(listTableResultFile, listTableStdoutBuf.Bytes(), os.ModePerm)

	var tables []string
	if tables, err = readLines(listTableResultFile); err != nil {
		pp.Fatal(err)
	}
	pp.Print(tables)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

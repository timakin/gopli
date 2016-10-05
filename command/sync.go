package command

import (
	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
	"github.com/k0kubun/pp"
	"golang.org/x/crypto/ssh"
	"bytes"
	"io/ioutil"
	"os/user"
	"strings"
	"path/filepath"
	"log"
)

type tomlConfig struct {
	Database map[string]Database
	SSH      map[string]SSH
}

type Database struct {
	Host             string
	Port             int
	ManagementSystem string
	Name             string
	User             string
	Password         string
}

type SSH struct {
	Host string
	Port string
	User string
	Key  string
}

func CmdSync(c *cli.Context) {
	var tmlconf tomlConfig
	if _, err := toml.DecodeFile(c.String("config"), &tmlconf); err != nil {
		// TODO: pkg/errors
		pp.Print(err)
	}

	// fromDBConf := tmlconf.Database[c.String("from")]
	// toDBConf := tmlconf.Database[c.String("to")]
	fromSSHConf := tmlconf.SSH[c.String("from")]
	// toSSHConf := tmlconf.SSH[c.String("to")]

	usr, _ := user.Current()
	keypathString := strings.Replace(fromSSHConf.Key,  "~", usr.HomeDir, 1)
	keypath, _ := filepath.Abs(keypathString)
	pp.Print(keypath)
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
	conn, err := ssh.Dial("tcp", fromSSHConf.Host + ":" + fromSSHConf.Port, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	err = session.Run("ls -l")
	pp.Print(stdoutBuf.String())
}

package command

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
	"github.com/k0kubun/pp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/timakin/gopli/constants"
	. "github.com/timakin/gopli/lib"
)

var srcDBConf Database
var srcSSHConf SSH
var dstDBConf Database
var dstSSHConf SSH

type tomlConfig struct {
	Database map[string]Database
	SSH      map[string]SSH
}

// Database settings
type Database struct {
	Host             string
	ManagementSystem string
	Name             string
	User             string
	Password         string
	Offset           int
}

// SSH settings
type SSH struct {
	Host string
	Port string
	User string
	Key  string
}

var listTableResultFile string
var loadDirName string
var srcHostConn *ssh.Client
var dstHostConn *ssh.Client
var tableBlackList = [3]string{"schema_migrations", "repli_chk", "repli_clock"}

const (
	DefaultOffset = 1000000000
)

// CmdSync supports `sync` command in CLI
func CmdSync(c *cli.Context) {
	setupMultiCore()
	loadTomlConf(c)
	connectToSrcHost()
	defer srcHostConn.Close()
	fetchTableList(srcHostConn)
	defer DeleteTmpDir(loadDirName)
	fetchTables(srcHostConn)
	connectToDstHost()
	if dstHostConn != nil {
		defer dstHostConn.Close()
	}
	deleteTables(dstHostConn)
	loadInfile(dstHostConn)
}

func setupMultiCore() {
	maxProcs := os.Getenv("GOMAXPROCS")

	if maxProcs == "" {
		cpus := runtime.NumCPU()
		runtime.GOMAXPROCS(cpus)
	}
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
		if isInBlackList(scanner.Text()) {
			continue
		}
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func isInBlackList(table string) bool {
	for _, blackListElem := range tableBlackList {
		if blackListElem == table {
			return true
		}
	}
	return false
}

func loadTomlConf(c *cli.Context) {
	log.Print("[Setting] loading toml configuration...")
	var tmlconf tomlConfig
	if _, err := toml.DecodeFile(c.String("config"), &tmlconf); err != nil {
		pp.Print(err)
	}

	srcDBConf = tmlconf.Database[c.String("from")]
	dstDBConf = tmlconf.Database[c.String("to")]
	srcSSHConf = tmlconf.SSH[c.String("from")]
	dstSSHConf = tmlconf.SSH[c.String("to")]
	log.Print("[Setting] loaded toml configuration")
}

func loadSrcSSHConf() *ssh.ClientConfig {
	usr, _ := user.Current()
	keypathString := strings.Replace(srcSSHConf.Key, "~", usr.HomeDir, 1)
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
		User: srcSSHConf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	return config
}

func connectToSrcHost() {
	config := loadSrcSSHConf()
	conn, err := ssh.Dial("tcp", srcSSHConf.Host+":"+srcSSHConf.Port, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	srcHostConn = conn
}

func fetchTableList(conn *ssh.Client) {
	log.Print("[Fetch] fetching the list of tables...")
	session, err := conn.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var listTableStdoutBuf bytes.Buffer
	session.Stdout = &listTableStdoutBuf
	listTableCmd := fmt.Sprintf(ShowTableCmd, srcDBConf.Name, srcDBConf.User, srcDBConf.Password)
	err = session.Run(listTableCmd)

	syncTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	loadDirName = "/tmp/db_sync_" + syncTimestamp
	if err := os.MkdirAll(loadDirName, 0777); err != nil {
		pp.Fatal(err)
	}

	listTableResultFile = loadDirName + "/" + srcDBConf.Name + "_list.txt"
	ioutil.WriteFile(listTableResultFile, listTableStdoutBuf.Bytes(), os.ModePerm)
	log.Print("[Fetch] completed fetching the list of tables")
}

func fetchTables(conn *ssh.Client) {
	log.Print("\t[Fetch] start to fetch table contents...")
	var tables []string
	tables, err := readLines(listTableResultFile)
	if err != nil {
		pp.Fatal(err)
	}

	sem := make(chan int, MaxFetchSession)
	var wg sync.WaitGroup
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			sem <- 1
			defer wg.Done()
			defer func() { <-sem }()
			session, err := conn.NewSession()
			if err != nil {
				panic("Failed to create session: " + err.Error())
			}
			defer session.Close()

			var fetchTableStdoutBuf bytes.Buffer
			session.Stdout = &fetchTableStdoutBuf
			fetchRowsCmd := fmt.Sprintf(SelectTablesCmd, srcDBConf.User, srcDBConf.Password, srcDBConf.Name, table)
			log.Print("\t\t[Fetch] fetcing " + table)
			err = session.Run(fetchRowsCmd)
			if err != nil {
				pp.Fatal(err)
			}
			fetchTableRowsResultFile := loadDirName + "/" + srcDBConf.Name + "_" + table + ".txt"
			ioutil.WriteFile(fetchTableRowsResultFile, fetchTableStdoutBuf.Bytes(), os.ModePerm)
			log.Print("\t\t[Fetch] completed fetcing " + table)
		}(table)
	}
	wg.Wait()
	log.Print("\t[Fetch] completed fetching all tables")
}

func loadDstSSHConf() *ssh.ClientConfig {
	if dstSSHConf.Host == "localhost" || dstSSHConf.Host == "127.0.0.1" {
		return nil
	}
	usr, _ := user.Current()
	keypathString := strings.Replace(dstSSHConf.Key, "~", usr.HomeDir, 1)
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
		User: dstSSHConf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	return config
}

func connectToDstHost() {
	config := loadDstSSHConf()
	if dstSSHConf.Host == "localhost" || dstSSHConf.Host == "127.0.0.1" {
		dstHostConn = nil
		return
	}
	conn, err := ssh.Dial("tcp", dstSSHConf.Host+":"+dstSSHConf.Port, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	dstHostConn = conn
}

func deleteTables(conn *ssh.Client) {
	log.Print("[Delete] deleting existing tables...")
	var tables []string
	tables, err := readLines(listTableResultFile)
	if err != nil {
		pp.Fatal(err)
	}

	sem := make(chan int, 5)
	var wg sync.WaitGroup
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			sem <- 1
			defer wg.Done()
			defer func() { <-sem }()

			log.Print("\t[Delete] deleting " + table)

			if dstSSHConf.Host == "localhost" || dstSSHConf.Host == "127.0.0.1" {
				var deleteTableCmd *exec.Cmd
				query := fmt.Sprintf(DeleteTableQuery, dstDBConf.Name, table)
				userOption := "-u" + dstDBConf.User
				executeOption := "--execute=" + query
				var passwordOption string
				if len(dstDBConf.Password) > 0 {
					passwordOption = "-p" + dstDBConf.Password
				} else {
					passwordOption = ""
				}
				deleteTableCmd = exec.Command("mysql", userOption, passwordOption, executeOption)

				err := deleteTableCmd.Run()
				if err != nil {
					pp.Fatal(err)
				}
			} else {
				var deleteTableCmd string
				if len(dstDBConf.Password) > 0 {
					deleteTableCmd = fmt.Sprintf(DeleteTableCmd, dstDBConf.User, dstDBConf.Password, dstDBConf.Name, table)
				} else {
					deleteTableCmd = fmt.Sprintf(DeleteTableCmdWithoutPass, dstDBConf.User, dstDBConf.Name, table)
				}

				var deleteTableStdoutBuf bytes.Buffer
				session, err := conn.NewSession()
				if err != nil {
					panic("Failed to create session: " + err.Error())
				}
				defer session.Close()
				session.Stdout = &deleteTableStdoutBuf
				err = session.Run(deleteTableCmd)
				if err != nil {
					pp.Fatal(err)
				}
			}
		}(table)
	}
	wg.Wait()
	log.Print("[Delete] completed deleting tables")
}

func loadInfile(conn *ssh.Client) {
	log.Print("[Load Infile] start to send fetched contents...")
	var tables []string
	tables, err := readLines(listTableResultFile)
	if err != nil {
		pp.Fatal(err)
	}
	sem := make(chan int, MaxLoadInfileSession)
	var wg sync.WaitGroup
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			sem <- 1
			defer wg.Done()
			defer func() { <-sem }()
			fetchedTableFile := loadDirName + "/" + srcDBConf.Name + "_" + table + ".txt"
			query := fmt.Sprintf(LoadInfileQuery, fetchedTableFile, dstDBConf.Name, table)
			var passwordOption string
			if len(dstDBConf.Password) > 0 {
				passwordOption = fmt.Sprintf("-p%s", dstDBConf.Password)
			} else {
				passwordOption = ""
			}
			log.Print("\t[Load Infile] start to send the contents inside of " + table)
			var cmd *exec.Cmd
			if dstSSHConf.Host == "localhost" || dstSSHConf.Host == "127.0.0.1" {
				cmd = exec.Command("mysql", "-u"+dstDBConf.User, passwordOption, "--enable-local-infile", "--execute="+query)
			} else {
				cmd = exec.Command("mysql", "-u"+dstDBConf.User, passwordOption, "-h"+dstSSHConf.Host, "--enable-local-infile", "--execute="+query)
			}
			err := cmd.Run()
			if err != nil {
				pp.Fatal(err)
			}
			log.Print("\t[Load Infile] completed sending the contents inside of " + table)
		}(table)
		wg.Wait()
	}
	log.Print("[Load Infile] completed sending fetched contents")
	log.Print("[Finished] All tasks finished")
}

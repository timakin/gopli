package database

import (
	. "github.com/timakin/gopli/constants"
	. "github.com/timakin/gopli/lib"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"
)

type DBFetcher interface {
	Fetch() error
}

type DBInserter interface {
	Clean() error
	Insert() error
}

type DBConnector struct {
	SSHClient        *ssh.Client
	Host             string
	ManagementSystem string
	Name             string
	User             string
	Password         string
	IsContainer      bool
}

func CreateFetcher(dbConf Database, sshConf SSH) (fetcher DBFetcher, err error) {
	// Connect to the host of the data soruce.
	config := LoadSrcSSHConf(sshConf.User, sshConf.Key)
	srcHostConn, err := ssh.Dial("tcp", sshConf.Host+":"+sshConf.Port, config)
	if err != nil {
		return nil, err
	}

	switch dbConf.ManagementSystem {
	case "mysql":
		return &MySQLFetcher{
			SSHClient:   srcHostConn,
			Host:        dbConf.Host,
			Name:        dbConf.Name,
			User:        dbConf.User,
			Password:    dbConf.Password,
			IsContainer: dbConf.IsContainer,
		}, nil
	default:
		return nil, nil
	}
}

func CreateInserter(dbConf Database, sshConf SSH) (inserter DBInserter, err error) {
	config, err := generateSSHSign(sshConf)
	if err != nil {
		return nil, err
	}
	var dstHostConn *ssh.Client
	if sshConf.Host == "localhost" || sshConf.Host == "127.0.0.1" {
		dstHostConn = nil
	} else {
		dstHostConn, err = ssh.Dial("tcp", sshConf.Host+":"+sshConf.Port, config)
		if err != nil {
			return nil, err
		}
	}

	switch dbConf.ManagementSystem {
	case "mysql":
		return &MySQLInserter{
			SSHClient:   dstHostConn,
			Host:        dbConf.Host,
			Name:        dbConf.Name,
			User:        dbConf.User,
			Password:    dbConf.Password,
			IsContainer: dbConf.IsContainer,
		}, nil
	default:
		return nil, nil
	}
}

func generateSSHSign(sshConf SSH) (*ssh.ClientConfig, error) {
	if sshConf.Host == "localhost" || sshConf.Host == "127.0.0.1" {
		return nil, nil
	}
	usr, _ := user.Current()
	keypathString := strings.Replace(sshConf.Key, "~", usr.HomeDir, 1)
	keypath, _ := filepath.Abs(keypathString)
	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: sshConf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	return config, nil
}

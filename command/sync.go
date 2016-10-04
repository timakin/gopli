package command

import (
	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
	"github.com/k0kubun/pp"
	"golang.org/x/crypto/ssh"
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
	User string
	Key  string
}

func CmdSync(c *cli.Context) {
	pp.Print(c.String("config"))
	var tmlconf tomlConfig
	if _, err := toml.DecodeFile(c.String("config"), &tmlconf); err != nil {
		// TODO: pkg/errors
		pp.Print(err)
	}
	pp.Print(tmlconf)

	usr, _ := user.Current()
	// usr.HomeDir = ホームディレクトリ
	keypathString := strings.Replace(tmlconf.SSH[c.String("from")].Key,  "~", usr.HomeDir, 1)
	keypath, _ := filepath.Abs(keypathString)
	pp.Print(keypath)
	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "user",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
	}
	conn, err := ssh.Dial("tcp", tmlconf.SSH[c.String("from")].Host, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()
}

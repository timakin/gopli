package lib

import (
	"golang.org/x/crypto/ssh"

	"os/user"
	"path/filepath"

	"io/ioutil"

	"log"
	"strings"
)

func LoadSrcSSHConf(sshUser string, keypath string) *ssh.ClientConfig {
	usr, _ := user.Current()
	keypath = strings.Replace(keypath, "~", usr.HomeDir, 1)
	absKeyPath, _ := filepath.Abs(keypath)
	key, err := ioutil.ReadFile(absKeyPath)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	return config
}

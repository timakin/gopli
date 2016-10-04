package command

import (
	"github.com/codegangsta/cli"
	"github.com/BurntSushi/toml"
	"github.com/k0kubun/pp"
)

type tomlConfig struct {
    Database map[string]database
    SSH map[string]ssh
}

type database struct {
    Host string
    Port int
    ManagementSystem string
		Name string
		User string
    Password string
}

type ssh struct {
    Key string
}

func CmdSync(c *cli.Context) {
	pp.Print(c.String("config"))
	var tmlconf tomlConfig
	if _, err := toml.DecodeFile(c.String("config"), &tmlconf); err != nil {
		// TODO: pkg/errors
		pp.Print(err)
	}
	pp.Print(tmlconf)
}

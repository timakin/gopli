package lib

import (
	"log"

	"github.com/BurntSushi/toml"
	"github.com/k0kubun/pp"
	. "github.com/timakin/gopli/constants"
)

type tomlConfig struct {
	Database map[string]Database
	SSH      map[string]SSH
}

func LoadTomlConf(configPath string) tomlConfig {
	log.Print("[Setting] loading toml configuration...")
	var tmlconf tomlConfig
	if _, err := toml.DecodeFile(configPath, &tmlconf); err != nil {
		pp.Print(err)
	}

	log.Print("[Setting] loaded toml configuration")
	return tmlconf
}

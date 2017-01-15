package command

import (
	"github.com/codegangsta/cli"
	. "github.com/timakin/gopli/constants"
	database "github.com/timakin/gopli/database"
	. "github.com/timakin/gopli/lib"
)

// CmdSync supports `sync` command in CLI
func CmdSync(c *cli.Context) {
	// Enable multi core setting
	SetupMultiCore()

	// Load tomlConfig
	tmlconf := LoadTomlConf(c.String("config"))

	// Create DB Fetcher
	fetcher, err := database.CreateFetcher(tmlconf.Database[c.String("from")], tmlconf.SSH[c.String("from")])
	if err != nil {
		panic("Failed to create fetcher instance: " + err.Error())
	}

	defer DeleteTmpDir(TMP_DIR_PATH)

	// Fetch
	err = fetcher.Fetch()
	if err != nil {
		panic("Failed to fetch: " + err.Error())
	}

	// Create DB Inserter
	inserter, err := database.CreateInserter(tmlconf.Database[c.String("to")], tmlconf.SSH[c.String("to")])
	if err != nil {
		panic("Failed to create inserter instance: " + err.Error())
	}

	// Clean up
	err = inserter.Clean()
	if err != nil {
		panic("Failed to clean: " + err.Error())
	}

	// INSERT
	err = inserter.Insert()
	if err != nil {
		panic("Failed to insert: " + err.Error())
	}
}

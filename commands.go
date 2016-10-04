package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/timakin/cylinder/command"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{
	{
		Name:   "sync",
		Usage:  "",
		Action: command.CmdSync,
		Flags:  []cli.Flag{
			cli.StringFlag{
				Name:  "config, c",
				Usage: "Load configuration from `FILE`",
			},
			cli.StringFlag{
				Name:  "from, f",
				Usage: "Target `HOST` for fetching data source",
			},
			cli.StringFlag{
				Name:  "to, t",
				Usage: "Target `HOST` to apply copied data from other host",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

package main

import (
	"fmt"
	"os"

	"github.com/sampointer/kapo/command"
	"gopkg.in/urfave/cli.v1"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{
	{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "run a command and close the socket on exit",
		Action:  command.CmdRun,
		Flags:   []cli.Flag{},
	},
	{
		Name:    "supervise",
		Aliases: []string{"s"},
		Usage:   "run and restart a command continually",
		Action:  command.CmdSupervise,
		Flags:   []cli.Flag{},
	},
	{
		Name:    "watch",
		Aliases: []string{"w"},
		Usage:   "report status of an externally invoked process",
		Action:  command.CmdWatch,
		Flags:   []cli.Flag{},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.\n", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

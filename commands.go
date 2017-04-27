package main

import (
	"fmt"
	"os"

	"github.com/sampointer/kapo/command"
	"gopkg.in/urfave/cli.v1"
)

//GlobalFlags defines global flags
var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "port, p",
		Value:  "6666",
		Usage:  "port to listen on `PORT`",
		EnvVar: "KAPO_PORT",
	},
	cli.StringFlag{
		Name:   "interface, i",
		Value:  "0.0.0.0",
		Usage:  "bind to interface `IP`",
		EnvVar: "KAPO_INTERFACE",
	},
}

//Commands defines subcommands
var Commands = []cli.Command{
	{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "run a command and close the socket on exit",
		Action:  command.CmdRun,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:   "ttl, t",
				Value:  0,
				Usage:  "Stop execution after `SECOND` seconds",
				EnvVar: "KAPO_TTL",
			},
		},
	},
	{
		Name:    "supervise",
		Aliases: []string{"s"},
		Usage:   "run and restart a command continually",
		Action:  command.CmdSupervise,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:   "wait, w",
				Value:  5,
				Usage:  "seconds to wait between restarts",
				EnvVar: "KAPO_WAIT",
			},
		},
	},
	{
		Name:    "watch",
		Aliases: []string{"w"},
		Usage:   "report status of an externally invoked process",
		Action:  command.CmdWatch,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:   "wait, w",
				Value:  5,
				Usage:  "seconds to wait between evaluating process list",
				EnvVar: "KAPO_WAIT",
			},
			cli.IntFlag{
				Name:   "pid, p",
				Value:  0,
				Usage:  "limit watched process to a single pid",
				EnvVar: "KAPO_WATCHPID",
			},
		},
	},
}

//CommandNotFound is invoked when an invalid subcommand is passed
func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.\n", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

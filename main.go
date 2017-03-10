package main

import (
	"gopkg.in/urfave/cli.v1"
	"os"
)

func main() {

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "Copyright 2017 Sam Pointer"
	app.Email = "sam@outsidethe.net"
	app.Usage = "Wrap any command in a status socket"

	app.Flags = GlobalFlags
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Run(os.Args)
}

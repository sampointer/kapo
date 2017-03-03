package command

import (
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
)

func CmdRun(c *cli.Context) error {

	process.Run(c)
	return nil
}

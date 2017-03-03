package command

import (
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
)

func CmdRun(c *cli.Context) error {

	status := process.Status{
		Command:   c.Args().First(),
		Arguments: c.Args().Tail(),
		Status:    "running",
		Mode:      "run",
		TTL:       time.Duration(c.Int("ttl")),
	}

	process.Setup(c, &status)

	status.StartTime = time.Now()
	err := process.Run(c)

	return nil
}

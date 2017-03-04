package command

import (
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
	"time"
)

func CmdSupervise(c *cli.Context) error {

	wait := time.Duration(c.Int("wait")) * time.Second

	status := process.Status{
		Command:   c.Args().First(),
		Arguments: c.Args().Tail(),
		Mode:      "supervise",
		Wait:      wait,
	}

	process.Setup(c, &status)

	for {
		status.StartTime = time.Now()
		status.Status = "running"
		status.ExitCode = 0
		rc, exit := process.Run(c)
		status.ExitCode = rc
		status.Status = exit

		time.Sleep(wait)
	}

	return nil
}

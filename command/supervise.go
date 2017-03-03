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
		Status:    "running",
		Mode:      "supervise",
	}

	process.Setup(c, &status)

	for {
		status.StartTime = time.Now()
		err := process.Run(c)

		if err != nil {
			status.ExitCode = err
			status.Status = "stopped"
		}

		time.Sleep(wait)
	}

	return nil
}

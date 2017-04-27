package command

import (
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
	"time"
)

//CmdRun runs a process until exit
func CmdRun(c *cli.Context) error {

	var statuses []process.Status

	status := process.Status{
		Command:   c.Args().First(),
		Arguments: c.Args().Tail(),
		Status:    "running",
		Mode:      "run",
		TTL:       time.Duration(c.Int("ttl")),
	}

	statuses = append(statuses, status)

	process.Setup(c, &statuses)

	statuses[0].StartTime = time.Now()
	_, _ = process.Run(c, "running")

	return nil
}

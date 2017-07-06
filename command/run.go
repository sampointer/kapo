package command

import (
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
	"time"
)

//CmdRun runs a process until exit
func CmdRun(c *cli.Context) error {

	var statuses []process.Status
	wait := time.Duration(c.Int("wait")) * time.Second

	status := process.Status{
		Command:   c.Args().First(),
		Arguments: c.Args().Tail(),
		Status:    "running",
		Mode:      "run",
		TTL:       time.Duration(c.Int("ttl")),
		Wait:      wait,
	}

	statuses = append(statuses, status)

	sidebindPort, _ := process.Setup(c, &statuses)

	statuses[0].SidebindPort = sidebindPort
	statuses[0].StartTime = time.Now()
	statuses[0].ExitCode, statuses[0].Status = process.Run(c, "running")
	statuses[0].EndTime = time.Now()

	time.Sleep(wait)

	return nil
}

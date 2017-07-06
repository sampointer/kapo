package command

import (
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
	"time"
)

//CmdSupervise runs a process and restart it upon failure
func CmdSupervise(c *cli.Context) error {

	var statuses []process.Status
	wait := time.Duration(c.Int("wait")) * time.Second

	status := process.Status{
		Command:   c.Args().First(),
		Arguments: c.Args().Tail(),
		Mode:      "supervise",
		Wait:      wait,
	}

	statuses = append(statuses, status)

	sidebindPort, _ := process.Setup(c, &statuses)

	for {
		statuses[0].SidebindPort = sidebindPort
		statuses[0].StartTime = time.Now()
		statuses[0].Status = "running"
		statuses[0].ExitCode = 0
		rc, exit := process.Run(c, "supervising")
		statuses[0].EndTime = time.Now()
		statuses[0].ExitCode = rc
		statuses[0].Status = exit

		time.Sleep(wait)
	}

}

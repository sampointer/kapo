package command

import (
	"github.com/sampointer/go-ps"
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
	"log"
)

func CmdWatch(c *cli.Context) error {

	var status process.Status
	var statuses []process.Status
	var watched []process.Status

	process.Setup(c, &statuses)

	for {
		// Get all processes
		procs, err := ps.Processes()
		if err != nil {
			log.Fatalf("Unable to obtain process list: %s", err)
		}

		watched = nil
		for _, p := range procs {
			// Get matching processes
			if p.Executable() == c.Args().First() {
				status = process.Status{
					Command:   c.Args().First(),
					Arguments: c.Args().Tail(),
					Mode:      "supervise",
				}

				watched = append(watched, status)
			}

		}

		statuses = watched

	}
	return nil
}

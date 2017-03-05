package command

import (
	"github.com/mitchellh/go-ps"
	"github.com/sampointer/kapo/process"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

func CmdWatch(c *cli.Context) error {

	var status process.Status
	var statuses []process.Status
	var watched []process.Status

	wait := time.Duration(c.Int("wait")) * time.Second
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
					Wait:      wait,
					StartTime: getstarttime(p.Pid()),
				}

				watched = append(watched, status)
			}

		}

		statuses = watched
		time.Sleep(wait)

	}
	return nil
}

func getstarttime(pid int) time.Time {
	var blanktime time.Time

	proc_path := path.Join("/proc", strconv.Itoa(pid))
	info, err := os.Stat(proc_path)

	if err != nil {
		return blanktime
	} else {
		return info.ModTime()
	}
}

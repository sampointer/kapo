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

	var proc_status string
	var start_time time.Time
	var status process.Status
	var statuses []process.Status
	var watched []process.Status

	wait := time.Duration(c.Int("wait")) * time.Second
	process.Setup(c, &statuses)

	if c.Int("pid") > 0 {
		log.Printf("Watching process %s with PID %d", c.Args().First(), c.Int("pid"))
	} else {
		log.Printf("Watching process %s without explicit PID", c.Args().First())
	}

	for {
		// If we're passed --pid find that explicit process
		if c.Int("pid") > 0 {
			proc, _ := ps.FindProcess(c.Int("pid"))

			if proc == nil {
				proc_status = "stopped"
			} else {
				proc_status = "running"
				start_time = getstarttime(c.Int("pid"))
			}

			status = process.Status{
				Command:   c.Args().First(),
				Mode:      "watch",
				StartTime: start_time,
				Wait:      wait,
				Status:    proc_status,
			}

			watched = nil
			watched = append(watched, status)
		} else {
			// Get all processes
			procs, err := ps.Processes()
			if err != nil {
				log.Fatalf("Unable to obtain process list: %s", err)
			} else {
				watched = nil
				for _, p := range procs {
					// Get matching processes
					if p.Executable() == c.Args().First() {
						status = process.Status{
							Command:   c.Args().First(),
							Mode:      "watch",
							StartTime: getstarttime(p.Pid()),
							Wait:      wait,
							Status:    "running",
						}

						watched = append(watched, status)
					}

				}
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

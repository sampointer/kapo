package command

import (
	"github.com/mitchellh/go-ps"
	"github.com/sampointer/kapo/process"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path"
	"strconv"
	"time"
)

//CmdWatch looks at the state of a process in the process list
func CmdWatch(c *cli.Context) error {

	var status process.Status
	var statuses []process.Status
	var watched []process.Status

	wait := time.Duration(c.Int("wait")) * time.Second
	sidebindPort, _ := process.Setup(c, &statuses)

	if c.Int("pid") > 0 {
		log.Printf("Watching process %s with PID %d", c.Args().First(), c.Int("pid"))
	} else {
		log.Printf("Watching process %s without explicit PID", c.Args().First())
	}

	for {
		// If we're passed --pid find that explicit process
		if c.Int("pid") > 0 {
			status = process.Status{
				Command:      c.Args().First(),
				Mode:         "watch",
				SidebindPort: sidebindPort,
				Wait:         wait,
			}

			proc, _ := ps.FindProcess(c.Int("pid"))

			if proc == nil {
				status.Status = "stopped"
			} else {
				status.Status = "running"
				status.StartTime = getstarttime(c.Int("pid"))
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
							Command:      c.Args().First(),
							Mode:         "watch",
							SidebindPort: sidebindPort,
							StartTime:    getstarttime(p.Pid()),
							Status:       "running",
							Wait:         wait,
						}

						watched = append(watched, status)
					}
				}
			}
		}

		if len(watched) > 0 {
			statuses = watched
		} else {
			status = process.Status{
				Command: c.Args().First(),
				Mode:    "watch",
				Wait:    wait,
				Status:  "stopped",
			}

			statuses = nil
			statuses = append(statuses, status)
		}

		time.Sleep(wait)

	}
}

func getstarttime(pid int) time.Time {
	var blanktime time.Time

	procPath := path.Join("/proc", strconv.Itoa(pid))
	info, err := os.Stat(procPath)

	if err != nil {
		return blanktime
	}
	return info.ModTime()
}

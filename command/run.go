package command

import "gopkg.in/urfave/cli.v1"
import "log"
import "os/exec"

func CmdRun(c *cli.Context) error {
	path, err := exec.LookPath(c.Args().First())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("executing %s", path)
	cmd := exec.Command(path)
	err = cmd.Run()
	if err != nil {
		log.Printf("exited %s", err)
	} else {
		log.Printf("exited 0")
	}

	return nil
}

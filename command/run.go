package command

import "gopkg.in/urfave/cli.v1"
import "log"
import "os/exec"
import "strings"

func CmdRun(c *cli.Context) error {
	path, err := exec.LookPath(c.Args().First())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("executing %s %s", path, strings.Join(c.Args().Tail(), " "))
	cmd := exec.Command(path, c.Args().Tail()...)
	if err := cmd.Run(); err != nil {
		log.Print(err)
	} else {
		log.Print("exited status 0")
	}

	return nil
}

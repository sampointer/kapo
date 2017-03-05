package process

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Status struct {
	Arguments []string
	Command   string
	EndTime   time.Time
	ExitCode  int
	Mode      string
	StartTime time.Time
	Status    string
	TTL       time.Duration
	Wait      time.Duration
}

func Setup(c *cli.Context, s *[]Status) error {

	// Start the status server in a gorountine
	bindaddr := fmt.Sprintf("%s:%s", c.GlobalString("interface"), c.GlobalString("port"))
	log.Printf("binding to %s", bindaddr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handler(w, r, *s) })
	go http.ListenAndServe(bindaddr, nil)

	return nil
}

func Run(c *cli.Context, modeverb string) (int, string) {

	var ctx context.Context
	var cancel context.CancelFunc
	var exit string
	var rc int
	var ttl time.Duration

	// Ensure we can find our executable
	path, err := exec.LookPath(c.Args().First())
	if err != nil {
		log.Fatal(err)
	}

	// If we've been given a TTL execute with a context
	if c.Int("ttl") != 0 {
		ttl = time.Duration(c.Int("ttl")) * time.Second
		ctx, cancel = context.WithTimeout(context.Background(), ttl)
		defer cancel()
		log.Printf("stopping execution after %ss TTL expires", ttl)
	} else {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, path, c.Args().Tail()...)

	// Start execution
	err = cmd.Start()
	if err != nil {
		log.Fatal("failed to start %s %s", path, strings.Join(c.Args().Tail(), " "))
	}
	log.Printf("%s %s %s", modeverb, path, strings.Join(c.Args().Tail(), " "))

	// Report the supervised process's exit status
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// Non-zero exit code
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				rc = status.ExitStatus()
				exit = "stopped"
			}
		} else {
			rc = 0
			exit = "killed"
		}
	} else {
		// Process exited properly
		rc = 0
		exit = "stopped"
	}

	return rc, exit
}

func handler(w http.ResponseWriter, r *http.Request, status []Status) {
	js, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

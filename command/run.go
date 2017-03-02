package command

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type Status struct {
	Command   string
	Arguments []string
	StartTime time.Time
	TTL       time.Duration
	Status    string
	ExitCode  int
}

func CmdRun(c *cli.Context) error {

	var ctx context.Context
	var cancel context.CancelFunc
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
	log.Printf("executing %s %s", path, strings.Join(c.Args().Tail(), " "))

	// Prepare a static status object
	status := Status{
		Command:   c.Args().First(),
		Arguments: c.Args().Tail(),
		Status:    "running",
		StartTime: time.Now(),
		TTL:       ttl,
	}

	// Start the status server in a gorountine
	bindaddr := fmt.Sprintf("%s:%s", c.GlobalString("interface"), c.GlobalString("port"))
	log.Printf("binding to %s", bindaddr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handler(w, r, status) })
	go http.ListenAndServe(bindaddr, nil)

	// Report the supervised process's exit status
	if err := cmd.Wait(); err != nil {
		log.Print(err)
	} else {
		log.Print("exited status 0")
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request, status Status) {
	js, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

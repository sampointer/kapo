package command

import (
	"encoding/json"
	"gopkg.in/urfave/cli.v1"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type StatusResponse struct {
	Command   string
	Arguments []string
	Status    string
}

func CmdRun(c *cli.Context) error {
	path, err := exec.LookPath(c.Args().First())
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(path, c.Args().Tail()...)
	err = cmd.Start()
	if err != nil {
		log.Fatal("failed to start %s %s", path, strings.Join(c.Args().Tail(), " "))
	}
	log.Printf("executing %s %s", path, strings.Join(c.Args().Tail(), " "))

	http.HandleFunc("/", handler)
	go http.ListenAndServe("localhost:6666", nil)

	if err := cmd.Wait(); err != nil {
		log.Print(err)
	} else {
		log.Print("exited status 0")
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	status_response := StatusResponse{"sam", []string{"shit"}, "ok"}
	js, err := json.Marshal(status_response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

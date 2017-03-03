package command

import (
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

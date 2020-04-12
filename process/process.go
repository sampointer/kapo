package process

import (
	"bufio"
	"context"
	"encoding/json"
	"runtime"
	//Blank import just to gain the default internal metrics
	_ "expvar"
	"fmt"
	"github.com/coreos/go-systemd/activation"
	"github.com/coreos/go-systemd/daemon"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Status is the status of a process
type Status struct {
	Arguments    []string
	Command      string
	EndTime      time.Time
	ExitCode     int
	Mode         string
	SidebindPort uint16
	StartTime    time.Time
	Status       string
	TTL          time.Duration
	Wait         time.Duration
}

//Setup starts the HTTP listener
func Setup(c *cli.Context, s *[]Status) (uint16, error) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handler(w, r, *s) })

	// Start the status server in a gorountine
	if c.GlobalBool("socket-activation") {
		//systemd is not supported on MacOS
		if runtime.GOOS == "darwin" {
			log.Fatal("systemd socket activation is not supported on MacOS")
		}

		// Expect listeners from systemd socket activation
		listeners, err := activation.Listeners()

		if err != nil {
			panic(err)
		}

		if len(listeners) != 1 {
			panic("Unexpected number of socket activation fds")
		}

		log.Printf("using socket activation")
		go http.Serve(listeners[0], nil)
		daemon.SdNotify(false, "READY=1")
	} else {
		// Bind conventionally otherwise
		bindaddr := interfaceandport(c.GlobalString("interface"), uint16(c.GlobalInt64("port")))
		log.Printf("binding to %s", bindaddr)
		go http.ListenAndServe(bindaddr, nil)
	}

	// If we've been asked to sidebind locate the next highest
	// available port and bind a listener to that, too.
	if c.GlobalBool("sidebind") {
		sidebindPort := uint16(c.GlobalInt64("port"))
		for {
			sidebindPort++
			log.Printf("attempting sidebinding to %d", sidebindPort)
			listen, err := net.Listen("tcp", fmt.Sprintf(":%d", sidebindPort))
			if err == nil {
				// Available port found, bind to it
				listen.Close()
				bindaddr := interfaceandport(c.GlobalString("interface"), sidebindPort)
				go http.ListenAndServe(bindaddr, nil)
				log.Printf("sidebinding to %s", bindaddr)
				return sidebindPort, nil
			}
		}
	}

	return 0, nil
}

// Run invokes the process using a given mode
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
		log.Printf("stopping execution after %s TTL expires", ttl)
	} else {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, path, c.Args().Tail()...)

	// STDOUT and STDERR processing
	if c.GlobalBool("stdout") {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		log.Info("capturing process stdout")
		go standardOutput(stdout, c.GlobalBool("stdlog"))
	}

	if c.GlobalBool("stderr") {
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}

		log.Info("capturing process stderr")
		go standardOutput(stderr, c.GlobalBool("stdlog"))
	}

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
	var err error
	var js []byte

	if len(status) > 0 {
		js, err = json.Marshal(status)
	} else {
		js, err = []byte(""), nil
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Given an interface and a port, return interface:port
func interfaceandport(i string, p uint16) string {
	return fmt.Sprintf("%s:%s", i, strconv.FormatUint(uint64(p), 10))
}

// standardOutput handles processing of STDOUT and STDERR of the supervised
// process
func standardOutput(s io.ReadCloser, l bool) {
	var line string
	rd := bufio.NewReader(s)

	for {
		line, _ = rd.ReadString('\n')
		if l {
			log.Printf("%s", strings.TrimSuffix(line, "\n"))
		} else {
			fmt.Printf("%s", line)
		}
	}
}

# Kapo [![Circle CI](https://circleci.com/gh/sampointer/kapo.svg?style=shield)](https://circleci.com/gh/sampointer/kapo) [![Go Report Card](https://goreportcard.com/badge/github.com/sampointer/kapo)](https://goreportcard.com/report/github.com/sampointer/kapo) [![GoDoc](https://godoc.org/github.com/sampointer/kapo?status.svg)](https://godoc.org/github.com/sampointer/kapo) [![Join the chat at https://gitter.im/kapo_tool](https://badges.gitter.im/kapo_tool.svg)](https://gitter.im/kapo_tool/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link) <img align="right" src="/assets/ganster-small.png" alt="ganster-small" />
Wrap any command in a status socket.


## Description
Kapo is a swiss-army knife for integrating programs that do not have their own method of presenting their status over a network with those that need it.
Examples might be:

1. Allow queue workers to be monitored by your service discovery system
1. Allow random shell scripts to present status to your container scheduler
1. Abuse load balancers to route traffic based on an open port
1. Expose the running status of configuration management tools
1. Alert if a process fails too often
1. Start and monitor non-networked programs on-demand using [systemd socket activation](https://github.com/sampointer/kapo#socket-activation)

When a program is executed under `kapo` a JSON-speaking HTTP server is started and the state of the process is reported to whomever requests it.

This server will respond correctly to HEAD requests. The open socket will function as an indicator of aliveness. The body is a JSON document reflecting
process state.

Requests could be as simple as inferring the process is alive if a socket can be opened to as complex as parsing the JSON schema and performing
selective actions based on the state reported therein.

### Modes

`kapo` can be run in one of three modes: `run`, `supervise` and `watch`.

The first is the most useful as a container `ENTRYPOINT`, especially in tandem with the `--ttl` flag to inject some chaos.

The second will prop up a failing process by continually restarting it if it fails (with an optional wait interval), reporting interesting facts like the last return code and start time on the status listener.

The third, `watch`, is for use in tandem with your preferred process supervisor: it'll infer the state of the
process from the process list of the operating system. By default `watch` will match all processes in the process list that match the binary name
given as the argument. Passing `--pid` limits the scope to just that process.

The utility of all of this is probably best illustrated via some examples:

### Examples
#### Abuse an ELB to keep a pool of queue workers running
1. Create an ELB with no external ingress
1. Configure the ELB Health check to perform a TCP or HTTP check on port 6666 with failure criteria that suit your application
1. Create an auto-scaling group with `HealthCheckType: ELB`
1. Have each host start their worker under `kapo`:

```bash
$ kapo --interface 0.0.0.0 --port 6666 run -- ./worker.py --dowork myqueue
```

Should the worker on any given node die the ELB will instruct the ASG to kill the node and another will be provisioned to replace it.

A slightly less expensive variant that resurrects worker processes if it can is:

```bash
$ kapo supervise --wait 30 -- ./worker.py --dowork myqueue
```

A human or computer can query the state of workers:

```bash
$ curl http://localhost:6666 2>/dev/null
[{"Arguments":["--dowork", "myqueue"],"Command":"./worker.py","EndTime":"0001-01-01T00:00:00Z","ExitCode":0,"Mode":"supervise","StartTime":"0001-01-01T00:00:00Z","Status":"running","TTL":0,"Wait":0}]
```

#### As a Container `ENTRYPOINT` to inject random failure, exercise your scheduler's resilience and add TTLs to containers
As all good proponents of the [SRE model](https://landing.google.com/sre/book.html) know, forcing failures by periodically killing execution units,
forcing circuit breakers to fire, and reguarly refreshing your running environment are vital. `kapo` can be used as a container `ENTRYPOINT` to
force containers to have a random TTL:

```
FROM ubuntu:latest
COPY kapo kapo
RUN apt-get update && apt-get install stress
EXPOSE 6666
ENTRYPOINT /bin/bash -c './kapo --interface 0.0.0.0 --port 6666 run --ttl $(($RANDOM % 30 + 1)) -- stress -c 1'
```

### Expose Puppet run activity
```bash
$ nohup kapo watch puppet &
$ curl http://somehost:6666 2>/dev/null
[{"Arguments":null,"Command":"puppet","EndTime":"0001-01-01T00:00:00Z","ExitCode":0,"Mode":"watch","StartTime":"0001-01-01T00:00:00Z","Status":"stopped","TTL":0,"Wait":5000000000}]
$ sleep 300
$ curl http://somehost:6666 2>/dev/null
[{"Arguments":null,"Command":"puppet","EndTime":"0001-01-01T00:00:00Z","ExitCode":0,"Mode":"watch","StartTime":"2017-03-02T18:20:28.762060588Z","Status":"running","TTL":0,"Wait":5000000000}]
```

## Socket Activation
Kapo can listen for connections via [systemd socket activation](http://0pointer.de/blog/projects/socket-activation.html) by passing the global option `--socket-activation` (or setting `KAPO_SOCKET_ACTIVATION`) and configuring systemd as appropriate.

When `--socket-activation` is passed any configured interface or port is ignored for the purposes of binding. The `--sidebind` global option will attempt to bind a second listener to the incrementing values above `--port` on `--interface` until it is successful.

There are a number of interesting use-cases for this functionality, including but not limited to starting and inspecting the status of non-networked programs and scripts on-demand upon receipt of a TCP connection, using `run` mode and `--sidebind`.

```
# useful.service
[Unit]
Description=Most Useful Script
Requires=network.target
After=multi-user.target

[Service]
Type=notify # required to enable the HTTP handler to notify dbus
ExecStart=/usr/local/bin/kapo --socket-activation --sidebind run /usr/local/bin/useful.sh
NonBlocking=true

[Install]
WantedBy=multi-user.target
```

```
# useful.socket
[Socket]
ListenStream=0.0.0.0:6666

[Install]
WantedBy=sockets.target
```

```bash
$ systemctl enable useful.socket
$ systemctl start useful.socket
$ systemctl enable useful.service # But not started on boot
```

One can then `curl http://localhost:6666` to have systemd start an instance
of `useful.sh`. The connection to the socket will remain open for the duration
of the execution of the process. We'll receive a status structure:

```
[{"Arguments":null,"Command":"/usr/local/bin/useful.sh","EndTime":"0001-01-01T00:00:00Z","ExitCode":0,"Mode":"run","StartTime":"2017-03-02T18:20:28.762060588Z","Status":"running","TTL":0,"SidebindPort": 6667, "Wait":5000000000}]
```

If we wanted to get a further update on the execution status of this process we cannot, of course, `curl http://localhost:6666` as this was start _another_ instance of the service. Instead we can see from the initial call above that we have a `SidebindPort` returned on which `kapo` has started another listener, by virtue of us passing `--sidebind`. We can interrogate this to our heart's content. If we _were_ to start another instance of the service whilst the original one was still executing we'd be returned a different `SidebindPort` by that instance.

## Configuration
### Environment Variables
Switch arguments can be configured by setting an appropriately named environment
variable:

* `KAPO_PORT`
* `KAPO_INTERFACE`
* `KAPO_SIDEBIND`
* `KAPO_SOCKET_ACTIVATION`
* `KAPO_TTL`
* `KAPO_WAIT`
* `KAPO_WATCHPID`

## Expvar
The listener exposes basic runtime metrics via [expvar](https://golang.org/pkg/expvar/) for use with [expvarmon](https://github.com/divan/expvarmon).

## Install

* Binaries built for `amd_64`: [version 0.5.0][1]
* Debian package: [version 0.5.0][2]
* rpm package: [version 0.5.0][3]
* Go: `go install github.com/sampointer/kapo`

## Contribution

1. Fork ([https://github.com/sampointer/kapo/fork](https://github.com/sampointer/kapo/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Kapo?
Keep A Port Open.

## Author

[sampointer](https://github.com/sampointer)

[1]: https://89-83670015-gh.circle-artifacts.com/0/tmp/circle-artifacts.LchfjK7/kapo
[2]: https://89-83670015-gh.circle-artifacts.com/0/tmp/circle-artifacts.LchfjK7/kapo_0.5.0-89_amd64.deb
[3]: https://89-83670015-gh.circle-artifacts.com/0/tmp/circle-artifacts.LchfjK7/kapo-0.5.0-89.x86_64.rpm

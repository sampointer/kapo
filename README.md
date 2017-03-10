# Kapo [![Circle CI](https://circleci.com/gh/sampointer/kapo.svg?style=svg)](https://circleci.com/gh/sampointer/kapo) <img align="right" src="/assets/ganster-small.png" alt="ganster-small" />
Wrap any command in a status socket.


## Description
Kapo is a swiss-army knife for integrating programs that do not have their own method of presenting their status over a network with those that need it.
Examples might be:

1. Allow queue workers to be monitored by your service discovery system
1. Allow random shell scripts to present status to your container scheduler
1. Abuse load balancers to route traffic based on an open port
1. Expose the running status of configuration management tools
1. Alert if a process fails too often

When a program is executed under `kapo` a JSON-speaking HTTP server is started and the state of the process is reported to whomever requests it.

This server will respond correctly to HEAD requests. The open socket will function as an indicator of aliveness. The body is a JSON document reflecting
process state.

Requests could be as simple as inferring the process is alive if a socket can be opened to as complex as parsing the JSON schema and performing
selective actions based on the state reported therein.

### Modes

`kapo` can be run in one of three modes: `run`, `supervise` and `watch`.

The first is the most useful as a container `ENTRYPOINT`, especially in tandem with the `--ttl` flag to inject some chaos.

The second will prop up a
failing process by continually restarting it if it fails (with an optional wait interval), reporting interesting facts like the last return code
and start time on the status listener.

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

## Install

* Binaries built for `amd_64`: [version 0.1.1][1]
* Debian package: [version 0.1.1][2]
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

[1]: https://39-83670015-gh.circle-artifacts.com/0/tmp/circle-artifacts.O63Zk2s/kapo
[2]: https://39-83670015-gh.circle-artifacts.com/0/tmp/circle-artifacts.O63Zk2s/kapo_0.1.1-39_amd64.deb

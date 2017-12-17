# go-oom

A simple test case for OOM event dump within docker container.
Foundermental could be found [here](https://www.kernel.org/doc/gorman/html/understand/understand016.html) and [here](https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt).

# Plan A & B

Plan A is that we try to add cap `SYS_RAWIO`, and try to catch the SIGTERM signal, as Linux kernel said that it would
send SIGTERM towards the killing process to wait it clean exit, more details could be found at [kernel doc Out Of Memory Management: 13.4  Killing the Selected Process](https://www.kernel.org/doc/gorman/html/understand/understand016.html)

Plan B is that we try to register kernel event, and record `memory.pressure_level` relative event, more details could be
found at [cgroup doc 10. OOM Control and the tail example script](https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt).

## test!

```
# build oom test image
docker build -t go-oom .

# limit with 20Mb, and container would be OOM killed!
# For Plan A(failed!): now, tested container is killed via -9, we should add SIGTERM support for docker!
# For Plan B(success!): using event tracker, we can receive the events!
docker run -v /sys/fs/cgroup:/sys/fs/cgroup --cap-add SYS_RAWIO -m 20m go-oom
```

## result for Plan A

Em.. saddly to see that container process was killed by `kill -9`, and the container exit status shown below:

```
...

"Created": "2017-12-17T10:02:19.005289788Z",
"Path": "/bin/sh",
"Args": [
    "-c",
    "/go-oom"
],
"State": {
    "Status": "exited",
    "Running": false,
    "Paused": false,
    "Restarting": false,
    "OOMKilled": true,
    "Dead": false,
    "Pid": 0,
    "ExitCode": 137,
    "Error": "",
    "StartedAt": "2017-12-17T10:02:19.250479381Z",
    "FinishedAt": "2017-12-17T10:02:19.406659477Z"
},

...
```

container logs here:

```
...
already used  18012160  bytes!
Killed  // directly be killed...
```

## result for Plan B

similiar to Plan A result, the ONLY different thing was that, we actually received events from kernel,
and thereafter, we can do sth within this goroutine!

```
2017/12/17 20:38:38 received event '\x02'
2017/12/17 20:38:38 received event '\x01'
2017/12/17 20:38:38 received event '\x01'
2017/12/17 20:38:38 received event '\x01'
2017/12/17 20:38:38 received event '\x01'
2017/12/17 20:38:38 received event '\x01'
2017/12/17 20:38:38 already used  17536000  bytes!
2017/12/17 20:38:38 already used  17561600  bytes!
2017/12/17 20:38:38 already used  17587200  bytes!
2017/12/17 20:38:38 received event '\x01'
2017/12/17 20:38:38 received event '\x01'
2017/12/17 20:38:38 received event '\x01'
```

# go-oom

A simple test case for OOM event dump within docker container.
Foundermental could be found [here](https://www.kernel.org/doc/gorman/html/understand/understand016.html).

## test!

```
# build oom test image
docker build -t oom-test .

# limit with 20Mb, and container would be OOM killed!
# TODO: now, tested container is killed via -9, we should add SIGTERM support for docker!
docker run --cap-add SYS_RAWIO -m 20m oom-test
```

## result

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

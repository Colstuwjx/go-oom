package main

/*
#include <assert.h>
#include <err.h>
#include <errno.h>
#include <fcntl.h>
#include <libgen.h>
#include <limits.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#include <sys/eventfd.h>

#define USAGE_STR "Usage: cgroup_event_listener <path-to-control-file> <args>"

// NOTE(Colstuwjx): sourcecode from https://github.com/torvalds/linux/blob/master/tools/cgroup/cgroup_event_listener.c
// I just changed `main` to `listen`, and make `efd` as an argument then removed `int efd = -1`
// and add some helper functions.

// C helper functions:
static char** makeCharArray(int size) {
        return calloc(sizeof(char*), size);
}

static void setArrayString(char **a, char *s, int n) {
        a[n] = s;
}

static void freeCharArray(char **a, int size) {
        int i;
        for (i = 0; i < size; i++)
                free(a[i]);
        free(a);
}

int listen(int argc, int efd, char **argv)
{
    int cfd = -1;
    int event_control = -1;
    char event_control_path[PATH_MAX];
    char line[LINE_MAX];
    int ret;

    if (argc != 3)
        errx(1, "%s", USAGE_STR);

    cfd = open(argv[1], O_RDONLY);
    if (cfd == -1)
        err(1, "Cannot open %s", argv[1]);

    ret = snprintf(event_control_path, PATH_MAX, "%s/cgroup.event_control",
            dirname(argv[1]));
    if (ret >= PATH_MAX)
        errx(1, "Path to cgroup.event_control is too long");

    event_control = open(event_control_path, O_WRONLY);
    if (event_control == -1)
        err(1, "Cannot open %s", event_control_path);

    //efd = eventfd(0, 0);

    if (efd == -1)
        err(1, "eventfd() failed");

    ret = snprintf(line, LINE_MAX, "%d %d %s", efd, cfd, argv[2]);
    if (ret >= LINE_MAX)
        errx(1, "Arguments string is too long");

    ret = write(event_control, line, strlen(line) + 1);
    if (ret == -1)
        err(1, "Cannot write to cgroup.event_control");

    while (1) {
        uint64_t result;

        ret = read(efd, &result, sizeof(result));
        if (ret == -1) {
            if (errno == EINTR)
                continue;
            err(1, "Cannot read from eventfd");
        }
        assert(ret == sizeof(result));

        ret = access(event_control_path, W_OK);
        if ((ret == -1) && (errno == ENOENT)) {
            puts("The cgroup seems to have removed.");
            break;
        }

        if (ret == -1)
            err(1, "cgroup.event_control is not accessible any more");

        printf("%s %s: crossed\n", argv[1], argv[2]);
    }

    return 0;
}
*/
import "C"

import (
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/sahne/eventfd"
)

func perform_fork() {
	var (
		// malloc_size    = 1024 * 1024 // bytes
		// print_interval = 100
		// num_chunks     = 20 * 1024
		malloc_size    = 1024 // bytes
		print_interval = 100
		num_chunks     = 20 * 1024
		chunk          []byte
		chunks         [][]byte
	)

	chunk = make([]byte, malloc_size)
	for i := 0; i <= malloc_size-1; i++ {
		chunk[i] = byte('x')
	}

	for i := 0; i <= num_chunks-1; i++ {
		if (i*malloc_size)%print_interval == 0 {
			log.Println("already used ", i*malloc_size, " bytes!")
		}

		tmp := make([]byte, malloc_size)
		copy(tmp, chunk)
		chunks = append(chunks, tmp)
	}

	log.Println("Done with ", num_chunks*malloc_size, " bytes!")
	log.Println("Entered in dead loop, please make me release memory by `Ctrl+C`!")

	for {
		time.Sleep(1 * time.Second)
	}
}

func watchEventFd(ready chan bool) {
	/*
		# an example from https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt
		# cd /sys/fs/cgroup/memory/
		# mkdir foo
		# cd foo
		# cgroup_event_listener memory.pressure_level low &
		# echo 8000000 > memory.limit_in_bytes
		# echo 8000000 > memory.memsw.limit_in_bytes
		# echo $$ > tasks
		# dd if=/dev/zero | read x

		(Expect a bunch of notifications, and eventually, the oom-killer will
		trigger.)
	*/

	// PLAN B: try to register kernel event via eventfd
	efd, err := eventfd.New()
	if err != nil {
		log.Fatalf("Could not create EventFD: %v", err)
	}

	// add event listener.
	// make char** arguments to pass control fd.
	sargs := []string{
		"cgroup_event_listener",
		"/sys/fs/cgroup/memory/memory.pressure_level",
		"low",
	}
	cargs := C.makeCharArray(C.int(len(sargs)))
	defer C.freeCharArray(cargs, C.int(len(sargs)))
	for i, s := range sargs {
		C.setArrayString(cargs, C.CString(s), C.int(i))
	}

	go func() {
		C.listen(C.int(3), C.int(efd.Fd()), cargs)
	}()

	ready <- true

	/* listen for new events */
	for {
		val, err := efd.ReadEvents()
		if err != nil {
			log.Printf("Error while reading from eventfd: %v", err)
			break
		}

		log.Printf("received event %q", val) // or do sth here, such as print stack.
	}
}

func main() {
	// PLAN A: test failed, OOM killer send `kill -9`...
	sig := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-sig:
				log.Println("Got it!")
				debug.PrintStack()

				time.Sleep(10 * time.Second)
				log.Println("Dumped!")
				done <- true
			}
		}
	}()

	// PLAN B: watch kernel event.
	eventReady := make(chan bool, 1)
	go watchEventFd(eventReady)
	<-eventReady

	go perform_fork()

	<-done
	log.Println("exiting")
}

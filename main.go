package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

func perform_fork() {
	var (
		// malloc_size    = 1024 * 1024 // bytes
		// print_interval = 100
		// num_chunks     = 20 * 1024
		malloc_size    = 1024 // bytes
		print_interval = 10
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
			fmt.Println("already used ", i*malloc_size, " bytes!")
		}

		tmp := make([]byte, malloc_size)
		copy(tmp, chunk)
		chunks = append(chunks, tmp)
	}

	fmt.Println("Done with ", num_chunks*malloc_size, " bytes!")
	fmt.Println("Entered in dead loop, please make me release memory by `Ctrl+C`!")

	for {
		time.Sleep(1 * time.Second)
	}
}

func main() {
	sig := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go perform_fork()
	go func() {
		for {
			select {
			case <-sig:
				fmt.Println("Got it!")
				debug.PrintStack()

				time.Sleep(10 * time.Second)
				fmt.Println("Dumped!")
				done <- true
			}
		}
	}()

	<-done
	fmt.Println("exiting")
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	versionString = "undefined"
)

func main() {
	var preStartCmd string
	var mainCmd string
	var postStopCmd string
	var version bool

	flag.StringVar(&preStartCmd, "pre", "", "Pre-start command")
	flag.StringVar(&mainCmd, "main", "", "Main command")
	flag.StringVar(&postStopCmd, "post", "", "Post-stop command")
	flag.BoolVar(&version, "version", false, "Display go-init version")
	flag.Parse()

	if version {
		fmt.Println(versionString)
		os.Exit(0)
	}

	if mainCmd == "" {
		log.Fatal("[go-init] No main command defined, exiting")
	}

	// Routine to reap zombies (it's the job of init)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go removeZombies(ctx, &wg)

	// Launch pre-start command
	if preStartCmd == "" {
		log.Println("[go-init] No pre-start command defined, skip")
	} else {
		log.Printf("[go-init] Pre-start command launched : %s\n", preStartCmd)
		err := run(preStartCmd)
		if err != nil {
			log.Println("[go-init] Pre-start command failed")
			log.Printf("[go-init] %s\n", err)
			cleanQuit(cancel, &wg, 1)
		} else {
			log.Printf("[go-init] Pre-start command exited")
		}
	}

	// Launch main command
	var mainRC int
	log.Printf("[go-init] Main command launched : %s\n", mainCmd)
	err := run(mainCmd)
	if err != nil {
		log.Println("[go-init] Main command failed")
		log.Printf("[go-init] %s\n", err)
		mainRC = 1
	} else {
		log.Printf("[go-init] Main command exited")
	}

	// Launch post-stop command
	if postStopCmd == "" {
		log.Println("[go-init] No post-stop command defined, skip")
	} else {
		log.Printf("[go-init] Post-stop command launched : %s\n", postStopCmd)
		err := run(postStopCmd)
		if err != nil {
			log.Println("[go-init] Post-stop command failed")
			log.Printf("[go-init] %s\n", err)
			cleanQuit(cancel, &wg, 1)
		} else {
			log.Printf("[go-init] Post-stop command exited")
		}
	}

	// Wait removeZombies goroutine
	cleanQuit(cancel, &wg, mainRC)
}

func removeZombies(ctx context.Context, wg *sync.WaitGroup) {
	for {
		var status syscall.WaitStatus

		// Wait for orphaned zombie process
		pid, _ := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)

		if pid <= 0 {
			// PID is 0 or -1 if no child waiting
			// so we wait for 1 second for next check
			time.Sleep(1 * time.Second)
		} else {
			// PID is > 0 if a child was reaped
			// we immediately check if another one
			// is waiting
			continue
		}

		// Non-blocking test
		// if context is done
		select {
		case <-ctx.Done():
			// Context is done
			// so we stop goroutine
			wg.Done()
			return
		default:
		}
	}
}

func run(command string) error {

	var commandStr string
	var argsSlice []string

	// Split cmd and args
	commandSlice := strings.Fields(command)
	commandStr = commandSlice[0]
	// if there is args
	if len(commandSlice) > 1 {
		argsSlice = commandSlice[1:]
	}

	// Register chan to receive system signals
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs)
	defer signal.Reset()

	// Define command and rebind
	// stdout and stdin
	cmd := exec.Command(commandStr, argsSlice...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Create a dedicated pidgroup
	// used to forward signals to
	// main process and all children
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Goroutine for signals forwarding
	go func() {
		for sig := range sigs {
			// Ignore SIGCHLD signals since
			// thez are only usefull for go-init
			if sig != syscall.SIGCHLD {
				// Forward signal to main process and all children
				syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
			}
		}
	}()

	// Start defined command
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait for command to exit
	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func cleanQuit(cancel context.CancelFunc, wg *sync.WaitGroup, code int) {
	// Signal zombie goroutine to stop
	// and wait for it to release waitgroup
	cancel()
	wg.Wait()

	os.Exit(code)
}

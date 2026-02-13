package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/sspencer/cron"
)

func runCmd(spec, cmd string) *cron.Job {
	c, err := cron.Run(spec, func() {
		fmt.Printf("Run command: %q\n", cmd)
		// Use sh -c to properly handle commands with arguments and shell features
		exe := exec.Command("sh", "-c", cmd)
		exe.Stdout = os.Stdout
		exe.Stderr = os.Stderr
		if err := exe.Run(); err != nil {
			fmt.Println("Error executing command:", err)
		}

		fmt.Println("----------------------------")
	})

	if err != nil {
		log.Fatal(err)
	}

	return c
}

func runTicker(spec string) *cron.Job {
	tick := true
	c, err := cron.Run(spec, func() {
		now := time.Now()
		t := now.Format("15:04:05.000")
		if tick {
			fmt.Printf("tick: %s\n", t)
		} else {
			fmt.Printf("TOCK: %s\n", t)
		}

		tick = !tick
	})

	if err != nil {
		log.Fatal(err)
	}

	return c
}

func main() {
	var c *cron.Job
	switch len(os.Args) {
	case 1:
		fmt.Println("Running ticker every minute:")
		c = runTicker("* * * * *")
	case 2:
		fmt.Println("Running ticker per input")
		c = runTicker(os.Args[1])
	default:
		c = runCmd(os.Args[1], os.Args[2])
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	c.Stop()
	fmt.Println("<ctrl-c> pressed, exiting")
}

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
		exe := exec.Command(cmd)
		exe.Stdout = os.Stdout
		exe.Stderr = os.Stderr
		err := exe.Run()
		if err != nil {
			fmt.Println("Error executing command:", err)
			os.Exit(1)
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
	args := os.Args
	n := len(args)

	var c *cron.Job
	if n == 1 {
		fmt.Println("Running ticker every minute:")
		c = runTicker("* * * * *")
	} else if n == 2 {
		fmt.Println("Running ticker per input")
		spec := args[1]
		c = runTicker(spec)
	} else {
		spec := args[1]
		cmd := args[2]

		c = runCmd(spec, cmd)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	c.Stop()
	fmt.Println("<ctrl-c> pressed, exiting")
}

package main

import (
	"flag"
	"fmt"
	"time"
)

func init() {
	cmd := &command{
		short: "prints the status of the current task",
		long:  "foob",
		usage: "status",

		needsBackend: true,

		flags: flag.NewFlagSet("start", flag.ExitOnError),
		run:   status,
	}

	commands["status"] = cmd
}

func status(c *command) {
	args := c.flags.Args()
	if len(args) != 0 {
		c.Usage(1)
	}

	log, err := defaultBackend.Status()
	if err != nil {
		c.Error(err)
	}
	if log == nil {
		fmt.Println("not working on any task")
		return
	}

	task, err := defaultBackend.Load(log.Name)
	if err != nil {
		c.Error(err)
	}

	fmt.Printf("working on %s since %s (%s)\n", task.Name, log.When, time.Since(log.When))
	fmt.Println(task)
}

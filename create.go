package main

import (
	"flag"
	"fmt"
	"time"
)

func init() {
	cmd := &command{
		short: "creates, adds the estimate, and starts a task",
		long:  "foob",
		usage: "create <task name> <estimate>",

		needsBackend: true,

		flags: flag.NewFlagSet("create", flag.ExitOnError),
		run:   create,
	}

	commands["create"] = cmd
}

func create(c *command) {
	args := c.flags.Args()
	if len(args) != 2 {
		c.Usage(1)
	}
	dur, err := time.ParseDuration(args[1])
	if err != nil {
		c.Error(err)
	}
	if err := stopIfStarted(); err != nil {
		c.Error(err)
	}
	task := &Task{
		Name:        args[0],
		Annotations: []Annotation{{When: time.Now(), EstimateDelta: dur}},
		Estimate:    dur,
	}
	if err := defaultBackend.Save(task); err != nil {
		c.Error(err)
	}
	if err := defaultBackend.Start(task.Name); err != nil {
		c.Error(err)
	}
	fmt.Println("started working on", task.Name)
	fmt.Println(task)
}

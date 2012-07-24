package main

import (
	"flag"
	"fmt"
)

func init() {
	cmd := &command{
		short: "creates a new task",
		long:  "foob",
		usage: "new <task name>",

		needsBackend: true,

		flags: flag.NewFlagSet("new", flag.ExitOnError),
		run:   newTask,
	}

	commands["new"] = cmd
}

func newTask(c *command) {
	args := c.flags.Args()
	if len(args) != 1 {
		c.Usage(1)
	}
	task := Task{
		Name: args[0],
	}
	if err := defaultBackend.Save(&task); err != nil {
		c.Error(err)
	}
	fmt.Println("created task:", task.Name)
}

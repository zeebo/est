package main

import (
	"flag"
	"fmt"
)

func init() {
	cmd := &command{
		short: "renames a task",
		long:  "gsafdg",
		usage: "mv <old> <new>",

		needsBackend: true,

		flags: flag.NewFlagSet("mv", flag.ExitOnError),
		run:   mv,
	}

	commands["mv"] = cmd
}

func mv(c *command) {
	args := c.flags.Args()
	if len(args) != 2 {
		c.Usage(1)
	}
	task, err := defaultBackend.Load(args[0])
	if err != nil {
		c.Error(err)
	}

	if err := defaultBackend.Rename(task.Name, args[1]); err != nil {
		c.Error(err)
	}

	fmt.Printf("moved %s to %s\n", args[0], args[1])
}

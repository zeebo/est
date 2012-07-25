package main

import (
	"flag"
	"fmt"
)

func init() {
	cmd := &command{
		short: "removes a task",
		long:  "gsafdg",
		usage: "rm <name>",

		needsBackend: true,

		flags: flag.NewFlagSet("rm", flag.ExitOnError),
		run:   rm,
	}

	commands["rm"] = cmd
}

func rm(c *command) {
	args := c.flags.Args()
	if len(args) != 1 {
		c.Usage(1)
	}
	name := args[0]

	task, err := defaultBackend.Load(name)
	if err != nil {
		c.Error(err)
	}

	if err := defaultBackend.Remove(name); err != nil {
		c.Error(err)
	}

	fmt.Printf("deleted %s\n", name)
	fmt.Println(task)
}

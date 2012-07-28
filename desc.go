package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func init() {
	cmd := &command{
		short: "sets the description for the task",
		long:  "foob",
		usage: "desc <task name> [description ...]",

		needsBackend: true,

		flags: flag.NewFlagSet("desc", flag.ExitOnError),
		run:   desc,
	}

	commands["desc"] = cmd
}

func desc(c *command) {
	args := c.flags.Args()
	if len(args) < 1 {
		c.Usage(1)
	}
	task, err := defaultBackend.Load(args[0])
	if err != nil {
		c.Error(err)
	}

	desc := strings.TrimSpace(strings.Join(args[1:], " "))
	if desc == "" {
		fmt.Println("Type your description (Ctrl+D to end):")
		var buf bytes.Buffer
		_, err := io.Copy(&buf, os.Stdin)
		if err != nil {
			c.Error(err)
		}
		desc = strings.TrimSpace(buf.String())
	}

	if err := defaultBackend.SetDescription(task, desc); err != nil {
		c.Error(err)
	}

	task.Description = desc

	fmt.Println("description updated.")
	fmt.Println(task)
}

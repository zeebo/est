package main

import (
	"flag"
	"fmt"
	"time"
)

func init() {
	cmd := &command{
		short: "adds estimate time to task",
		long:  "gsafdg",
		usage: "add-est <task> <time>",

		needsBackend: true,

		flags: flag.NewFlagSet("add-est", flag.ExitOnError),
		run:   addEst,
	}

	commands["add-est"] = cmd
}

func addEst(c *command) {
	args := c.flags.Args()
	if len(args) != 2 {
		c.Usage(1)
	}
	task, err := defaultBackend.Load(args[0])
	if err != nil {
		c.Error(err)
	}
	dur, err := time.ParseDuration(args[1])
	if err != nil {
		c.Error(err)
	}

	ann := Annotation{
		When:          time.Now(),
		EstimateDelta: dur,
	}
	if err := defaultBackend.AddAnnotation(task.Name, ann); err != nil {
		c.Error(err)
	}
	task.Apply(ann)
	fmt.Println(task)
}

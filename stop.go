package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func init() {
	cmd := &command{
		short: "stops working on the current task",
		long:  "foob",
		usage: "stop",

		needsBackend: true,

		flags: flag.NewFlagSet("stop", flag.ExitOnError),
		run:   stop,
	}

	commands["stop"] = cmd
	commands["done"] = cmd //add done as an alias
}

func stop(c *command) {
	args := c.flags.Args()
	if len(args) != 0 {
		c.Usage(0)
	}

	log, err := defaultBackend.Status()
	if err != nil {
		c.Error(err)
	}
	if log == nil {
		fmt.Fprintln(os.Stdout, "not started on any task. use start first")
		os.Exit(1)
	}

	if err := defaultBackend.Stop(); err != nil {
		c.Error(err)
	}

	dur := time.Since(log.When)
	fmt.Println("adding", dur, "to", log.Name)

	task, err := defaultBackend.Load(log.Name)
	if err != nil {
		c.Error(err)
	}

	ann := Annotation{
		When:        time.Now(),
		ActualDelta: dur,
	}
	if err := defaultBackend.AddAnnotation(task, ann); err != nil {
		c.Error(err)
	}

	task.Apply(ann)
	fmt.Println(task)
}

package main

import (
	"flag"
	"fmt"
	"time"
)

func init() {
	cmd := &command{
		short: "starts working on a task",
		long:  "foob",
		usage: "start <task name>",

		needsBackend: true,

		flags: flag.NewFlagSet("start", flag.ExitOnError),
		run:   start,
	}

	commands["start"] = cmd
}

func start(c *command) {
	args := c.flags.Args()
	if len(args) != 1 {
		c.Usage(1)
	}

	if err := stopIfStarted(); err != nil {
		c.Error(err)
	}
	task, err := defaultBackend.Load(args[0])
	if err != nil {
		c.Error(err)
	}
	if err := defaultBackend.Start(task.Name); err != nil {
		c.Error(err)
	}

	fmt.Println("started working on", task.Name)
	fmt.Println(task)
}

func stopIfStarted() (err error) {
	log, err := defaultBackend.Status()
	if err != nil {
		return
	}
	if log == nil {
		return
	}
	fmt.Println("already working on", log.Name)
	if err = defaultBackend.Stop(); err != nil {
		return
	}
	dur := time.Since(log.When)
	fmt.Println("adding", dur, "to", log.Name)

	ann := Annotation{
		When:        time.Now(),
		ActualDelta: dur,
	}
	task, err := defaultBackend.Load(log.Name)
	if err != nil {
		return
	}
	if err = defaultBackend.AddAnnotation(task, ann); err != nil {
		return
	}
	return
}

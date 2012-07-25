package main

import (
	"flag"
	"fmt"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

func init() {
	cmd := &command{
		short: "adds actual time to task",
		long:  "gsafdg",
		usage: "add [-when=] <task> <time>",

		needsBackend: true,

		flags: flag.NewFlagSet("add", flag.ExitOnError),
		run:   add(makeActualAnno),
	}

	cmd.flags.StringVar(&addParams.addWhen, "when", "", "when the task should be added (default now). format: "+timeFormat)

	commands["add"] = cmd
}

func init() {
	cmd := &command{
		short: "adds estimate time to task",
		long:  "gsafdg",
		usage: "add-est [-when=] <task> <time>",

		needsBackend: true,

		flags: flag.NewFlagSet("add-est", flag.ExitOnError),
		run:   add(makeEstimateAnno),
	}

	cmd.flags.StringVar(&addParams.addWhen, "when", "", "when the task should be added (default now). format: "+timeFormat)

	commands["add-est"] = cmd
}

var addParams struct {
	addWhen string
}

func makeActualAnno(when time.Time, dur time.Duration) Annotation {
	return Annotation{
		When:        when,
		ActualDelta: dur,
	}
}

func makeEstimateAnno(when time.Time, dur time.Duration) Annotation {
	return Annotation{
		When:          when,
		EstimateDelta: dur,
	}
}

type annoMaker func(time.Time, time.Duration) Annotation

func ParseLocal(format, value string) (when time.Time, err error) {
	when, err = time.Parse(format, value)
	if err != nil {
		return
	}
	when = time.Date(
		when.Year(),
		when.Month(),
		when.Day(),
		when.Hour(),
		when.Minute(),
		when.Second(),
		when.Nanosecond(),
		time.Local,
	)
	return
}

func add(maker annoMaker) func(*command) {
	return func(c *command) {
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

		when := time.Now()
		if addParams.addWhen != "" {
			when, err = ParseLocal(timeFormat, addParams.addWhen)
			if err != nil {
				c.Error(err)
			}
			fmt.Println(when)
		}

		ann := maker(when, dur)
		if err := defaultBackend.AddAnnotation(task, ann); err != nil {
			c.Error(err)
		}
		task.Apply(ann)
		fmt.Println("added:", ann)
		fmt.Println(task)
	}
}

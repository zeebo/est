package main

import (
	"flag"
	"fmt"
	"os"
)

type command struct {
	short string //short description
	long  string //long help description
	usage string //usage

	needsBackend bool

	flags *flag.FlagSet
	run   func(*command)
}

func (c *command) Usage(status int) {
	fmt.Fprintln(os.Stderr, "usage:", c.usage)
	c.flags.PrintDefaults()
	os.Exit(status)
}

func (c *command) Error(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

var commands = map[string]*command{}

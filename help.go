package main

import (
	"flag"
	"fmt"
	"os"
)

func init() {
	cmd := &command{
		short: "displays help for commands",
		long:  "clever girl",
		usage: "help [topic]",

		needsBackend: false,

		flags: flag.NewFlagSet("help", flag.ExitOnError),
		run:   help,
	}

	commands["help"] = cmd
}

func help(c *command) {
	args := c.flags.Args()
	if len(args) == 0 {
		Usage(0)
	}

	topic := c.flags.Arg(0)
	if cmd, ok := commands[topic]; ok {
		fmt.Println("usage:", cmd.usage)
		cmd.flags.PrintDefaults()
		fmt.Println("")
		fmt.Println(cmd.long)
	} else {
		fmt.Fprintf(os.Stderr, "unknown command: %q\n", topic)
		c.Usage(1)
	}
}

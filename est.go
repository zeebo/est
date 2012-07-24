package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

var configPath string

func main() {
	const defaultPath = "~/.est"
	flag.StringVar(&configPath, "config", defaultPath, "path to configuration file")
	flag.Parse()

	//load the configuration if we can
	f, err := os.Open(configPath)
	if err != nil {
		//we had an error opening it. check if the path is the default path
		//in which case just silently use the default config
		if configPath == defaultPath {
			goto configLoaded
		}

		//print that we had an error opening it and use the default config
		fmt.Fprintf(os.Stderr, "%s.\nusing default configuration.\n\n", err)
		goto configLoaded
	}

	//attempt to load the configuration
	err = json.NewDecoder(f).Decode(defaultConfig)
	f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config file: %s\n", err)
		os.Exit(1)
	}

configLoaded:
	//grab the command out
	args := flag.Args()
	if len(args) == 0 {
		Usage(1)
	}

	cmdname := args[0]
	cmd, ok := commands[cmdname]
	if !ok {
		UnknownCommand(cmdname)
	}

	//connect to the backend only if the command needs it
	if cmd.needsBackend {
		if err := loadBackend(defaultConfig); err != nil {
			fmt.Fprintf(os.Stderr, "unable to connect to backend: %s\n", err)
			os.Exit(1)
		}
	}

	cmd.flags.Parse(args[1:])
	cmd.run(cmd)
}

func Usage(status int) {
	fmt.Fprintln(os.Stderr, "est is a tool for managing estimates\n")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "\test command [args]\n")
	fmt.Fprintln(os.Stderr, "The commands are:\n")

	names := make([]string, 0, len(commands))
	for k := range commands {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, cmdname := range names {
		cmd := commands[cmdname]
		fmt.Fprintf(os.Stderr, "\test % -10s %s\n", cmdname, cmd.short)
	}
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(status)
}

func UnknownCommand(name string) {
	fmt.Fprintf(os.Stderr, "est: unknown command %q\n", name)
	fmt.Fprintln(os.Stderr, "Run 'est help' for usage.")
	os.Exit(1)
}

package main

import (
	"encoding/json"
	"flag"
	"os"
)

type Config struct {
	Backend     string
	MongoConfig MongoConfig `json:",omitempty"`
}

type MongoConfig struct {
	Host       string `json:",omitempty"`
	Port       string `json:",omitempty"`
	Username   string `json:",omitempty"`
	Password   string `json:",omitempty"`
	Database   string `json:",omitempty"`
	Collection string `json:",omitempty"`
}

var defaultConfig = &Config{
	Backend: "mongo",
	MongoConfig: MongoConfig{
		Host:       "localhost",
		Database:   "est",
		Collection: "est",
	},
}

func init() {
	cmd := &command{
		short: "prints out the configuration",
		long:  "doofogy",
		usage: "config",

		needsBackend: false,

		flags: flag.NewFlagSet("config", flag.ExitOnError),
		run:   config,
	}

	commands["config"] = cmd
}

func config(c *command) {
	args := c.flags.Args()
	if len(args) != 0 {
		c.Usage(1)
	}
	b, _ := json.MarshalIndent(defaultConfig, "", "\t")
	os.Stdout.Write(b)
	os.Stdout.Write([]byte{'\n'})
}

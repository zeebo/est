package main

import (
	"fmt"
	"time"
)

type Backend interface {
	Save(task *Task) (err error)
	AddAnnotation(task *Task, a Annotation) (err error)
	PopAnnotation(task *Task) (err error)
	Load(name string) (task *Task, err error)
	Start(name string) (err error)
	Stop() (err error)
	Status() (log *StartLog, err error)
	Find(regex string, before, after time.Time) (tasks []*Task, err error)
	Rename(oldn, newn string) (err error)
	Remove(name string) (err error)
}

var defaultBackend Backend

func loadBackend(c *Config) (err error) {
	var b Backend
	switch c.Backend {
	case "mongo":
		b, err = openMongo(c.MongoConfig)
		if err == nil {
			defaultBackend = b
		}
	case "rpc":
		b, err = openRPC(c.RPCConfig)
		if err == nil {
			defaultBackend = b
		}
	default:
		err = fmt.Errorf("unknown backend: %q", c.Backend)
	}
	return
}

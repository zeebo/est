package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/url"
)

type Backend interface {
	Save(task *Task) (err error)
	AddAnnotation(name string, a Annotation) (err error)
	Load(name string) (task *Task, err error)
}

var defaultBackend Backend

type mongoBackend struct {
	C *mgo.Collection
}

type d map[string]interface{}

func (m *mongoBackend) Save(task *Task) (err error) {
	//while theres a task with this name, increment the number on the end of it
	candidate := task.Name
	for i := 1; ; i++ {
		var n int
		n, err = m.C.Find(d{"name": candidate}).Count()
		if err != nil {
			return
		}
		if n == 0 {
			task.Name = candidate
			break
		}
		candidate = fmt.Sprintf("%s%d", task.Name, i)
	}

	err = m.C.Insert(task)
	return
}

func (m *mongoBackend) Load(name string) (task *Task, err error) {
	task = new(Task)
	err = m.C.Find(d{"name": name}).One(task)
	return
}

func (m *mongoBackend) AddAnnotation(name string, a Annotation) (err error) {
	//create the change document
	ch := bson.D{
		{"$push", d{"annotations": a}},
		{"$inc", d{"estimate": a.EstimateDelta}},
		{"$inc", d{"actual": a.ActualDelta}},
	}
	err = m.C.Update(d{"name": name}, ch)
	return
}

func loadBackend(c *Config) (err error) {
	var b Backend
	switch c.Backend {
	case "mongo":
		b, err = openMongo(c.MongoConfig)
		if err == nil {
			defaultBackend = b
		}
	default:
		err = fmt.Errorf("unknown backend: %q", c.Backend)
	}
	return
}

func openMongo(c MongoConfig) (b Backend, err error) {
	//build a url for connecting based on the config
	u := &url.URL{
		Scheme: "mongodb",
		Host:   c.Host,
	}

	//only add credentials and database in the url if they're specified
	if c.Username != "" && c.Password == "" {
		u.User = url.User(c.Username)
		u.Path = "/" + c.Database
	}
	if c.Username != "" && c.Password != "" {
		u.User = url.UserPassword(c.Username, c.Password)
		u.Path = "/" + c.Database
	}

	s, err := mgo.Dial(u.String())
	if err != nil {
		err = fmt.Errorf("dial %s: %s", u, err)
		return
	}
	b = &mongoBackend{
		C: s.DB(c.Database).C(c.Collection),
	}
	return
}

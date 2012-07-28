package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/url"
	"time"
)

type MongoConfig struct {
	Host     string `json:",omitempty"`
	Port     string `json:",omitempty"`
	Username string `json:",omitempty"`
	Password string `json:",omitempty"`
	Database string `json:",omitempty"`
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
		tasks:    s.DB(c.Database).C("tasks"),
		startlog: s.DB(c.Database).C("startlog"),
	}
	return
}

type mongoBackend struct {
	tasks    *mgo.Collection
	startlog *mgo.Collection
}

type d map[string]interface{}

func (m *mongoBackend) Save(task *Task) (err error) {
	//while theres a task with this name, increment the number on the end of it
	candidate := task.Name
	for i := 1; ; i++ {
		var n int
		n, err = m.tasks.Find(d{"name": candidate}).Count()
		if err != nil {
			return
		}
		if n == 0 {
			task.Name = candidate
			break
		}
		candidate = fmt.Sprintf("%s%d", task.Name, i)
	}

	err = m.tasks.Insert(task)
	return
}

func (m *mongoBackend) Load(name string) (task *Task, err error) {
	task = new(Task)
	err = m.tasks.Find(d{"name": name}).One(task)
	return
}

func (m *mongoBackend) SetDescription(task *Task, desc string) (err error) {
	ch := bson.D{
		{"$set", d{"description": desc}},
	}
	err = m.tasks.Update(d{"name": task.Name}, ch)
	return
}

func (m *mongoBackend) AddAnnotation(task *Task, a Annotation) (err error) {
	//create the change document
	ch := bson.D{
		{"$push", d{"annotations": a}},
		{"$inc", d{"estimate": a.EstimateDelta}},
		{"$inc", d{"actual": a.ActualDelta}},
	}
	err = m.tasks.Update(d{"name": task.Name}, ch)
	return
}

func (m *mongoBackend) PopAnnotation(task *Task) (err error) {
	if len(task.Annotations) == 0 {
		err = fmt.Errorf("no annotations to undo")
	}

	//get the last annotation and negate it
	a := task.Annotations[len(task.Annotations)-1].Negate()

	//create the change document
	ch := bson.D{
		{"$pop", d{"annotations": 1}},
		{"$inc", d{"estimate": a.EstimateDelta}},
		{"$inc", d{"actual": a.ActualDelta}},
	}
	err = m.tasks.Update(d{"name": task.Name}, ch)
	return
}

func (m *mongoBackend) Rename(oldn, newn string) (err error) {
	ch := d{"$set": d{"name": newn}}
	err = m.tasks.Update(d{"name": oldn}, ch)
	return
}

func (m *mongoBackend) Remove(name string) (err error) {
	err = m.tasks.Remove(d{"name": name})
	return
}

func (m *mongoBackend) Start(name string) (err error) {
	err = m.startlog.Insert(StartLog{
		Name: name,
		When: time.Now(),
	})
	return
}

func (m *mongoBackend) Stop() (err error) {
	_, err = m.startlog.RemoveAll(nil)
	return
}

func (m *mongoBackend) Status() (log *StartLog, err error) {
	n, err := m.startlog.Count()
	if err != nil {
		return
	}
	if n == 0 {
		return
	}
	log = new(StartLog)
	err = m.startlog.Find(nil).One(log)
	return
}

func (m *mongoBackend) Find(regex string, before, after time.Time) (tasks []*Task, err error) {
	err = m.tasks.Find(d{
		"name":             d{"$regex": regex},
		"annotations.when": d{"$lt": after, "$gte": before},
	}).All(&tasks)
	return
}

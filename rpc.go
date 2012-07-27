package main

import (
	"encoding/gob"
	"fmt"
	"net/rpc"
	"time"
)

type RPCConfig struct {
	Network string `json:",omitempty"`
	Address string `json:",omitempty"`
}

func openRPC(c RPCConfig) (b Backend, err error) {
	cl, err := rpc.DialHTTP(c.Network, c.Address)
	if err != nil {
		return
	}
	b = &rpcClient{cl: cl}
	return
}

//
// handling error types: flatten to string
//

func init() {
	gob.Register(errString(""))
}

type errString string

func (e errString) Error() string {
	return string(e)
}

func wrapError(e *error) {
	if *e != nil {
		*e = errString(fmt.Sprint(*e))
	}
}

//
// rpcClient
//

type rpcClient struct {
	cl *rpc.Client
}

type None struct{}

var nul = new(None)

type RpcAddAnnotationArgs struct {
	Task *Task
	A    Annotation
}

type RpcFindArgs struct {
	Regex  string
	Before time.Time
	After  time.Time
}

type RpcRenameArgs struct {
	Oldn, Newn string
}

type RpcStatusReply struct {
	Log    *StartLog
	Exists bool
}

func (r *rpcClient) Save(task *Task) (err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.Save", task, nul)
	return
}

func (r *rpcClient) AddAnnotation(task *Task, a Annotation) (err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.AddAnnotation", RpcAddAnnotationArgs{
		Task: task,
		A:    a,
	}, nul)
	return
}

func (r *rpcClient) PopAnnotation(task *Task) (err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.PopAnnotation", task, nul)
	return
}

func (r *rpcClient) Load(name string) (task *Task, err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.Load", name, &task)
	return
}

func (r *rpcClient) Start(name string) (err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.Start", name, nul)
	return
}

func (r *rpcClient) Stop() (err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.Stop", nul, nul)
	return
}

func (r *rpcClient) Status() (log *StartLog, err error) {
	defer wrapError(&err)
	var reply RpcStatusReply
	err = r.cl.Call("Estimate.Status", nul, &reply)
	if reply.Exists {
		log = reply.Log
	}
	return
}

func (r *rpcClient) Find(regex string, before, after time.Time) (tasks []*Task, err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.Find", RpcFindArgs{
		Regex:  regex,
		Before: before,
		After:  after,
	}, &tasks)
	return
}

func (r *rpcClient) Rename(oldn, newn string) (err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.Rename", RpcRenameArgs{
		Oldn: oldn,
		Newn: newn,
	}, nul)
	return
}

func (r *rpcClient) Remove(name string) (err error) {
	defer wrapError(&err)
	err = r.cl.Call("Estimate.Remove", name, nul)
	return
}

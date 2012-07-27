package main

import (
	"flag"
	"net/http"
	"net/rpc"
)

func init() {
	cmd := &command{
		short: "serves rpc requests as a backend",
		long:  "dofasdf",
		usage: "serve <address>",

		needsBackend: true,

		flags: flag.NewFlagSet("serve", flag.ExitOnError),
		run:   serve,
	}

	commands["serve"] = cmd
}

func serve(c *command) {
	args := c.flags.Args()
	if len(args) == 0 {
		c.Usage(1)
	}
	http.ListenAndServe(args[0], nil)
}

//
// rpc server
//

func init() {
	if err := rpc.RegisterName("Estimate", rpcServer{}); err != nil {
		panic(err)
	}
	rpc.HandleHTTP()
}

type rpcServer struct{}

func (rpcServer) Save(task *Task, nul *None) (err error) {
	err = defaultBackend.Save(task)
	return
}

func (rpcServer) AddAnnotation(args *RpcAddAnnotationArgs, nul *None) (err error) {
	err = defaultBackend.AddAnnotation(args.Task, args.A)
	return
}

func (rpcServer) PopAnnotation(task *Task, nul *None) (err error) {
	err = defaultBackend.PopAnnotation(task)
	return
}

func (rpcServer) Load(name string, task **Task) (err error) {
	*task, err = defaultBackend.Load(name)
	return
}

func (rpcServer) Start(name string, nul *None) (err error) {
	err = defaultBackend.Start(name)
	return
}

func (rpcServer) Stop(nula *None, nulb *None) (err error) {
	err = defaultBackend.Stop()
	return
}

func (rpcServer) Status(nul *None, reply *RpcStatusReply) (err error) {
	log, err := defaultBackend.Status()

	//exsists is the assertion that the log is not nil
	reply.Exists = (log != nil)

	//if it doesn't exist, we need to make one to not send a nil pointer
	if !reply.Exists {
		reply.Log = new(StartLog)
	} else {
		reply.Log = log
	}
	return
}

func (rpcServer) Find(args *RpcFindArgs, tasks *[]*Task) (err error) {
	*tasks, err = defaultBackend.Find(args.Regex, args.Before, args.After)
	return
}

func (rpcServer) Rename(args *RpcRenameArgs, nul *None) (err error) {
	err = defaultBackend.Rename(args.Oldn, args.Newn)
	return
}

func (rpcServer) Remove(name string, nul *None) (err error) {
	err = defaultBackend.Remove(name)
	return
}

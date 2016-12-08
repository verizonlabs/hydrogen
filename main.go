package main

import (
	"sprint/scheduler"
	"sprint/executor"
	"github.com/verizonlabs/mesos-go/mesosproto"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"log"
	"os"
	"flag"
)

var (
	master = flag.String("master", "127.0.0.1:5050", "Master address <ip:port>")
	executorPath = flag.String("executor", "./example_executor", "Path to test executor.")
)

func init(){
	flag.Parse()
}

func main(){
	if *master == "127.0.0.1:5050" {
		*master = "172.17.8.101:5050"
	}

	// Make the scheduler
	sched := scheduler.NewScheduler(executor.NewExecInfo())

	info := mesosproto.FrameworkInfo{
		User: proto.String("tim"),
		Name: proto.String("test cluster"),
		Id: &mesosproto.FrameworkID{Value: proto.String("0123456")},

	}

	// scheduler, frameworkInfo, master string
	driver, err := scheduler.NewSchedulerDriver(sched, &info, *master)

	if err != nil {
		fmt.Printf("Scheduler Driver failed to be created: %v", err)
	}

	if status, err := driver.Start(); err != nil{
		log.Fatalf("Framework stopped with status %s and error: %s\n", status.String(), err.Error())
		os.Exit(-1)
	}
}

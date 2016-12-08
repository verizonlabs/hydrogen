package main

import (
	"sprint/scheduler"
	"sprint/executor"
	"github.com/verizonlabs/mesos-go/mesosproto"
	"fmt"
	"github.com/gogo/protobuf/proto"
)

func main(){
	master := "172.17.8.101:5050"

	// Make the scheduler
	sched := scheduler.NewScheduler(executor.NewExecInfo())

	info := mesosproto.FrameworkInfo{
		User: proto.String("tim"),
		Name: proto.String("test cluster"),
		Id: &mesosproto.FrameworkID{Value: proto.String("0123456")},

	}

	// scheduler, frameworkInfo, master string
	driver, err := scheduler.NewSchedulerDriver(sched, &info, master)

	if err != nil {
		fmt.Printf("Scheduler Driver failed to be created: %v", err)
	}

	driver.Start()
}

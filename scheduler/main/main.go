package main

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"log"
	client "mesos-framework-sdk/client"
	mesos "mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/scheduler"
	"sprint/scheduler"
	"sprint/scheduler/server"
	"sprint/scheduler/server/api"
	"sprint/scheduler/server/file"
	"time"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	srvConfig := new(server.ServerConfiguration).Initialize()
	schedulerConfig := new(scheduler.SchedulerConfiguration).Initialize().SetExecutorSrvCfg(srvConfig)

	executorSrv := file.NewExecutorServer(srvConfig)
	apiSrv := api.NewApiServer(srvConfig)

	// Parse here to catch flags defined in structures above.
	flag.Parse()

	frameworkInfo := &mesos.FrameworkInfo{
		User:            proto.String("root"),
		Name:            proto.String("Sprint"),
		FailoverTimeout: proto.Float64(5 * time.Second.Seconds()),
		Checkpoint:      proto.Bool(true),
		Role:            proto.String("*"),
		Hostname:        proto.String(""),
		Principal:       proto.String(""),
	}

	c := client.NewClient(schedulerConfig.Endpoint())
	s := sched.NewScheduler(c, frameworkInfo)

	log.Println("Starting executor server...")
	go executorSrv.Serve()
	log.Println("Starting API server...")
	go apiSrv.RunAPI()

}

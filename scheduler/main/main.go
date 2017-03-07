package main

import (
	"flag"
	"log"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources/manager"
	sched "mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	"mesos-framework-sdk/server/file"
	"mesos-framework-sdk/task_manager"
	"sprint/scheduler"
	"sprint/scheduler/api"
	"sprint/scheduler/events"
	"strings"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {

	// Executor/API server configuration.
	cert := flag.String("server.cert", "", "TLS certificate")
	key := flag.String("server.key", "", "TLS key")
	path := flag.String("server.executor.path", "executor", "Path to the executor binary")
	port := flag.Int("server.executor.port", 8081, "Executor server listen port")

	srvConfig := server.NewConfiguration(*cert, *key, *path, *port)
	schedulerConfig := new(scheduler.SchedulerConfiguration).Initialize()
	executorSrv := file.NewExecutorServer(srvConfig)
	apiSrv := api.NewApiServer(srvConfig)

	// Parse here to catch flags defined in structures above.
	flag.Parse()

	log.Println("Starting executor server...")
	go executorSrv.Serve()

	frameworkInfo := &mesos_v1.FrameworkInfo{
		User:            &schedulerConfig.User,
		Name:            &schedulerConfig.Name,
		FailoverTimeout: &schedulerConfig.Failover,
		Checkpoint:      &schedulerConfig.Checkpointing,
		Role:            &schedulerConfig.Role,
		Hostname:        &schedulerConfig.Hostname,
		Principal:       &schedulerConfig.Principal,
	}

	eventChan := make(chan *mesos_v1_scheduler.Event)

	kv := etcd.NewClient(strings.Split(schedulerConfig.StorageEndpoints, ","), schedulerConfig.StorageTimeout)
	m := task_manager.NewDefaultTaskManager()
	c := client.NewClient(schedulerConfig.MesosEndpoint)
	s := sched.NewDefaultScheduler(c, frameworkInfo)
	r := manager.NewDefaultResourceManager()
	e := eventcontroller.NewSprintEventController(s, m, r, eventChan, kv)

	log.Println("Starting API server...")
	go apiSrv.RunAPI(e)
	e.Run()

}

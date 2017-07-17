package main

import (
	"encoding/base64"
	"flag"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources/manager"
	sched "mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	"mesos-framework-sdk/server/file"
	t "mesos-framework-sdk/task/manager"
	"sprint/scheduler"
	"sprint/scheduler/api"
	apiManager "sprint/scheduler/api/manager"
	"sprint/scheduler/events"
	sprintTaskManager "sprint/task/manager"
	"sprint/task/persistence"
	"strings"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	logger := logging.NewDefaultLogger()

	// Define our framework here.
	config := new(scheduler.Configuration).Initialize()
	frameworkInfo := &mesos_v1.FrameworkInfo{
		User:            &config.Scheduler.User,
		Name:            &config.Scheduler.Name,
		FailoverTimeout: &config.Scheduler.Failover,
		Checkpoint:      &config.Scheduler.Checkpointing,
		Role:            &config.Scheduler.Role,
		Hostname:        &config.Scheduler.Hostname,
		Principal:       &config.Scheduler.Principal,
	}

	flag.Parse()

	// Executor Server
	execSrvCfg := server.NewConfiguration(
		config.FileServer.Cert,
		config.FileServer.Key,
		config.FileServer.Path,
		config.FileServer.Port,
	)
	executorSrv := file.NewExecutorServer(execSrvCfg, logger)

	// Executor server serves up our custom executor binary, if any.
	logger.Emit(logging.INFO, "Starting executor file server")
	go executorSrv.Serve()

	// Used to listen for events coming from mesos master to our scheduler.
	eventChan := make(chan *mesos_v1_scheduler.Event)

	// Storage interface that holds client and retry policy manager.
	p := persistence.NewPersistence(etcd.NewClient(
		strings.Split(config.Persistence.Endpoints, ","),
		config.Persistence.Timeout,
		config.Persistence.KeepaliveTime,
		config.Persistence.KeepaliveTimeout,
	), config)

	// Manages our tasks.
	t := sprintTaskManager.NewTaskManager(
		make(map[string]t.Task),
		p,
		config,
		logger,
	)

	auth := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(config.Scheduler.Principal+":"+config.Scheduler.Secret),
	)

	r := manager.NewDefaultResourceManager() // Manages resources from the cluster
	c := client.NewClient(client.ClientData{
		Endpoint: config.Scheduler.MesosEndpoint,
		Auth:     auth,
	}, logger) // Manages scheduler/executor HTTP calls, authorization, and new master detection.
	s := sched.NewDefaultScheduler(c, frameworkInfo, logger) // Manages how to route and schedule tasks.
	m := apiManager.NewApiParser(r, t, s)                    // Middleware for our API.

	// Event controller manages scheduler events and how they are handled.
	e := events.NewSprintEventController(config, s, t, r, eventChan, p, logger)

	logger.Emit(logging.INFO, "Starting API server")

	// Run our API in a go routine to listen for user requests.
	config.APIServer.Server = server.NewConfiguration(
		config.APIServer.Cert,
		config.APIServer.Key,
		"",
		config.APIServer.Port,
	)

	apiSrv := api.NewApiServer(config, m, logger)
	go apiSrv.RunAPI(nil) // nil means to use default handlers.

	// Run our event controller
	// Runs an election for the leader
	// and then subscribes to Mesos master to start listening for events.
	e.Run()
}

package main

import (
	"flag"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources/manager"
	sched "mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	"mesos-framework-sdk/server/file"
	"mesos-framework-sdk/structures"
	"net/http"
	"sprint/scheduler"
	"sprint/scheduler/api"
	"sprint/scheduler/events"
	sprintTaskManager "sprint/task/manager"
	"strings"
)

const (
	API_VERSION      = "v1"
	DEFAULT_MAP_SIZE = 100
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

	// Storage client
	kv := etcd.NewClient(
		strings.Split(config.Persistence.Endpoints, ","),
		config.Persistence.Timeout,
		config.Persistence.KeepaliveTime,
		config.Persistence.KeepaliveTimeout,
	)

	// Storage Engine
	engine := etcd.NewEtcdEngine(kv)

	// Wire up dependencies for the event controller
	t := sprintTaskManager.NewTaskManager(
		structures.NewConcurrentMap(DEFAULT_MAP_SIZE),
		engine,
		config,
		logger,
	) // Manages our tasks

	r := manager.NewDefaultResourceManager()                      // Manages resources from the cluster
	c := client.NewClient(config.Scheduler.MesosEndpoint, logger) // Manages HTTP calls
	s := sched.NewDefaultScheduler(c, frameworkInfo, logger)      // Manages how to route and schedule tasks.

	// Event controller manages scheduler events and how they are handled.
	e := events.NewSprintEventController(config, s, t, r, eventChan, engine, logger)

	logger.Emit(logging.INFO, "Starting API server")

	// Run our API in a go routine to listen for user requests.
	apiSrvCfg := server.NewConfiguration(
		config.APIServer.Cert,
		config.APIServer.Key,
		"",
		config.APIServer.Port,
	)

	apiSrv := api.NewApiServer(apiSrvCfg, s, t, r, http.NewServeMux(), API_VERSION, logger)
	go apiSrv.RunAPI(nil) // nil means to use default handlers.

	// Run our event controller
	// Runs an election for the leader
	// and then subscribes to mesos master to start listening for events.
	e.Run()
}

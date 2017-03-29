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
	"sprint/scheduler/ha"
	sprintTaskManager "sprint/task/manager"
	"strings"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	logger := logging.NewDefaultLogger()

	// Executor/API server configuration.
	cert := flag.String("server.cert", "", "TLS certificate")
	key := flag.String("server.key", "", "TLS key")
	path := flag.String("server.executor.path", "executor", "Path to the executor binary")
	port := flag.Int("server.executor.port", 8081, "Executor server listen port")
	apiPort := flag.Int("server.api.port", 8080, "API server listen port")

	// Define our framework here
	schedConfig := new(scheduler.SchedulerConfiguration).Initialize()
	frameworkInfo := &mesos_v1.FrameworkInfo{
		User:            &schedConfig.User,
		Name:            &schedConfig.Name,
		FailoverTimeout: &schedConfig.Failover,
		Checkpoint:      &schedConfig.Checkpointing,
		Role:            &schedConfig.Role,
		Hostname:        &schedConfig.Hostname,
		Principal:       &schedConfig.Principal,
	}

	flag.Parse()

	// Executor Server
	srvConfig := server.NewConfiguration(*cert, *key, *path, *port)
	executorSrv := file.NewExecutorServer(srvConfig, logger)

	// API server
	apiSrv := api.NewApiServer(srvConfig, http.NewServeMux(), apiPort, "v1", logger)

	logger.Emit(logging.INFO, "Starting executor file server")

	// Executor server serves up our custom executor binary, if any.
	go executorSrv.Serve()

	// Used to listen for events coming from mesos master to our scheduler.
	eventChan := make(chan *mesos_v1_scheduler.Event)

	// Storage client
	kv := etcd.NewClient(
		strings.Split(schedConfig.StorageEndpoints, ","),
		schedConfig.StorageTimeout,
		schedConfig.StorageKeepaliveTime,
		schedConfig.StorageKeepaliveTimeout,
	)
	// Storage Engine
	engine := etcd.NewEtcdEngine(kv)

	// Wire up dependencies for the event controller

	m := sprintTaskManager.NewTaskManager(structures.NewConcurrentMap(100)) // Manages our tasks
	r := manager.NewDefaultResourceManager()                                // Manages resources from the cluster
	c := client.NewClient(schedConfig.MesosEndpoint, logger)                // Manages HTTP calls
	s := sched.NewDefaultScheduler(c, frameworkInfo, logger)                // Manages how to route and schedule tasks.

	// Event controller manages scheduler events and how they are handled.
	e := events.NewSprintEventController(schedConfig, s, m, r, eventChan, engine, logger)

	logger.Emit(logging.INFO, "Starting leader election socket server")
	go ha.LeaderServer(schedConfig, logger)

	// Block here until we either become a leader or a standby.
	// If we are the leader we break out and continue to execute the rest of the scheduler.
	// If we are a standby then we connect to the leader and wait for the process to start over again.
	ha.LeaderElection(schedConfig, e, engine, logger)

	logger.Emit(logging.INFO, "Starting API server")

	// Run our API in a go routine to listen for user requests.
	go apiSrv.RunAPI(e, nil) // nil means to use default handlers.

	// Run our event controller to subscribe to mesos master and start listening for events.
	e.Run()
}

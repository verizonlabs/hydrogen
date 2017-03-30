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

const (
	API_VERSION      = "v1"
	DEFAULT_MAP_SIZE = 100
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
	srvConfig := server.NewConfiguration(*cert, *key, *path, *port)
	executorSrv := file.NewExecutorServer(srvConfig, logger)

	logger.Emit(logging.INFO, "Starting executor file server")

	// Executor server serves up our custom executor binary, if any.
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

	logger.Emit(logging.INFO, "Starting leader election socket server")
	go ha.LeaderServer(config, logger)

	// Block here until we either become a leader or a standby.
	// If we are the leader we break out and continue to execute the rest of the scheduler.
	// If we are a standby then we connect to the leader and wait for the process to start over again.
	ha.LeaderElection(config, e, engine, logger)

	logger.Emit(logging.INFO, "Starting API server")

	// Run our API in a go routine to listen for user requests.
	apiSrv := api.NewApiServer(srvConfig, s, t, r, http.NewServeMux(), apiPort, API_VERSION, logger)
	go apiSrv.RunAPI(nil) // nil means to use default handlers.

	// Run our event controller to subscribe to mesos master and start listening for events.
	e.Run()
}

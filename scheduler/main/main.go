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
	"mesos-framework-sdk/structures"
	"net/http"
	"sprint/scheduler"
	"sprint/scheduler/api"
	"sprint/scheduler/events"
	taskmanager "sprint/task/manager"
	"strings"
)

func CreateFrameworkInfo(config *scheduler.SchedulerConfiguration) *mesos_v1.FrameworkInfo {
	return &mesos_v1.FrameworkInfo{
		User:            &config.User,
		Name:            &config.Name,
		FailoverTimeout: &config.Failover,
		Checkpoint:      &config.Checkpointing,
		Role:            &config.Role,
		Hostname:        &config.Hostname,
		Principal:       &config.Principal,
	}
}

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {

	// Executor/API server configuration.
	cert := flag.String("server.cert", "", "TLS certificate")
	key := flag.String("server.key", "", "TLS key")
	path := flag.String("server.executor.path", "executor", "Path to the executor binary")
	port := flag.Int("server.executor.port", 8081, "Executor server listen port")

	// Executor Server
	srvConfig := server.NewConfiguration(*cert, *key, *path, *port)
	executorSrv := file.NewExecutorServer(srvConfig)

	// API server
	apiPort := flag.Int("server.api.port", 8080, "API server listen port")
	apiSrv := api.NewApiServer(srvConfig, http.NewServeMux(), apiPort, "v1")

	// Define our framework here
	schedulerConfig := new(scheduler.SchedulerConfiguration).Initialize()
	frameworkInfo := CreateFrameworkInfo(schedulerConfig)

	flag.Parse()

	// Executor server serves up our custom executor binary, if any.
	log.Println("Starting executor server...")
	go executorSrv.Serve()

	// Used to listen for events coming from mesos master to our scheduler.
	eventChan := make(chan *mesos_v1_scheduler.Event)

	// Wire up dependencies for the event controller
	kv := etcd.NewClient(
		strings.Split(schedulerConfig.StorageEndpoints, ","),
		schedulerConfig.StorageTimeout,
	) // Storage client
	m := taskmanager.NewTaskManager(structures.NewConcurrentMap(100)) // Manages our tasks
	r := manager.NewDefaultResourceManager()                          // Manages resources from the cluster
	c := client.NewClient(schedulerConfig.MesosEndpoint)              // Manages HTTP calls
	s := sched.NewDefaultScheduler(c, frameworkInfo)                  // Manages how to route and schedule tasks.
	// Event controller manages scheduler events and how they are handled.
	e := eventcontroller.NewSprintEventController(s, m, r, eventChan, kv)

	log.Println("Starting API server...")
	// Run our API in a go routine to listen for user requests.
	go apiSrv.RunAPI(e, nil) // nil means to use default handlers.
	// Run our event controller to subscribe to mesos master and start listening for events.
	e.Run()

}

package main

import (
	"flag"
	"github.com/golang/protobuf/proto"
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
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"net/http"
	"sprint/scheduler"
	"sprint/scheduler/api"
	"sprint/scheduler/events"
	sprintTaskManager "sprint/task/manager"
	"strings"
	"time"
)

// NOTE: This should be refactored out of the main file.
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

// NOTE: This should be in the event manager.
// Keep our state in check by periodically reconciling.
// This is recommended by Mesos.
func periodicReconcile(c *scheduler.SchedulerConfiguration, e *events.SprintEventController) {
	ticker := time.NewTicker(c.ReconcileInterval)

	for {
		select {
		case <-ticker.C:

			recon, err := e.TaskManager().GetState(sdkTaskManager.LAUNCHED)
			if err != nil {
				// log here.
				continue
			}
			e.Scheduler().Reconcile(recon)
		}
	}
}

// NOTE: This should be in the event manager.
// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
func restoreTasks(kv *etcd.Etcd, t *sprintTaskManager.SprintTaskManager) error {
	tasks, err := kv.ReadAll("/tasks")
	if err != nil {
		return err
	}

	for _, value := range tasks {
		var taskInfo mesos_v1.TaskInfo
		proto.UnmarshalText(value, &taskInfo)
		t.Add(&taskInfo)
	}

	return nil
}

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
	schedulerConfig := new(scheduler.SchedulerConfiguration).Initialize()
	frameworkInfo := CreateFrameworkInfo(schedulerConfig)

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

	// Wire up dependencies for the event controller
	kv := etcd.NewClient(
		strings.Split(schedulerConfig.StorageEndpoints, ","),
		schedulerConfig.StorageTimeout,
	) // Storage client
	m := sprintTaskManager.NewTaskManager(structures.NewConcurrentMap(100)) // Manages our tasks
	r := manager.NewDefaultResourceManager()                                // Manages resources from the cluster
	c := client.NewClient(schedulerConfig.MesosEndpoint, logger)            // Manages HTTP calls
	s := sched.NewDefaultScheduler(c, frameworkInfo, logger)                // Manages how to route and schedule tasks.

	// Event controller manages scheduler events and how they are handled.
	e := events.NewSprintEventController(s, m, r, eventChan, kv, logger)

	logger.Emit(logging.INFO, "Starting API server")

	// Run our API in a go routine to listen for user requests.
	go apiSrv.RunAPI(e, nil) // nil means to use default handlers.

	// Recover our state (if any) in the event we (or the server) go down.
	logger.Emit(logging.INFO, "Restoring any persisted state from data store")
	if err := restoreTasks(kv, m); err != nil {
		logger.Emit(logging.ERROR, "Failed to restore tasks from persistent data store")
	}

	// Kick off our scheduled reconciling.
	logger.Emit(logging.INFO, "Starting periodic reconciler thread with a %g minute interval", schedulerConfig.ReconcileInterval.Minutes())
	go periodicReconcile(schedulerConfig, e)

	// Run our event controller to subscribe to mesos master and start listening for events.
	e.Run()
}

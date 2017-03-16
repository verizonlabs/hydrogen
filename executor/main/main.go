package main

import (
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/include/executor"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"os"
	"sprint/executor"
	"sprint/executor/events"
	"strings"
	"time"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	logger := logging.NewDefaultLogger()
	// Add in environmental variables.
	var endpoint string
	var port string
	var storageEndpoint string
	//var storageTimeout string
	endpoint = os.Getenv("EXECUTOR_ENDPOINT")
	port = os.Getenv("EXECUTOR_ENDPOINT_PORT")
	storageEndpoint = os.Getenv("EXECUTOR_STORAGE_ENDPOINT")
	//storageTimeout := os.Getenv("EXECUTOR_STORAGE_TIMEOUT")

	if endpoint == "" {
		endpoint = "localhost"
	}
	if port == "" {
		port = "5050"
	}
	if storageEndpoint == "" {
		storageEndpoint = "localhost:2379"
	}
	etcdTimeout := 5 * time.Second
	kv := etcd.NewClient(
		strings.Split(storageEndpoint, ","),
		etcdTimeout,
	) // Storage client
	logger.Emit(logging.INFO, "Endpoint set to "+"http://"+endpoint+":"+port+"/api/v1/executor")
	c := client.NewClient("http://"+endpoint+":"+port+"/api/v1/executor", logger)
	ex := executor.NewSprintExecutor(nil, nil, c, logger)
	e := events.NewSprintExecutorEventController(ex, make(chan *mesos_v1_executor.Event), logger, kv)
	e.Run()
}

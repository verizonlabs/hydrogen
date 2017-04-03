package main

import (
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/executor"
	"mesos-framework-sdk/include/executor"
	"mesos-framework-sdk/logging"
	"os"
	"sprint/executor/events"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	logger := logging.NewDefaultLogger()
	endpoint := os.Getenv("EXECUTOR_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost"
	}

	port := os.Getenv("EXECUTOR_ENDPOINT_PORT")
	if port == "" {
		port = "5050"
	}

	storageEndpoint := os.Getenv("EXECUTOR_STORAGE_ENDPOINT")
	if storageEndpoint == "" {
		storageEndpoint = "localhost:2379"
	}

	logger.Emit(logging.INFO, "Endpoint set to "+"http://"+endpoint+":"+port+"/api/v1/executor")
	c := client.NewClient("http://"+endpoint+":"+port+"/api/v1/executor", logger)
	ex := executor.NewDefaultExecutor(nil, nil, c, logger)
	e := events.NewSprintExecutorEventController(ex, make(chan *mesos_v1_executor.Event), logger)
	e.Run()
}

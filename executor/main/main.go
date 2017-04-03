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

	storageTimeout, err := time.ParseDuration(os.Getenv("EXECUTOR_STORAGE_TIMEOUT"))
	if err != nil {
		logger.Emit(logging.ERROR, "Failed to parse storage timeout: %s", err.Error())
	}

	storageKaTime, err := time.ParseDuration(os.Getenv("EXECUTOR_STORAGE_KEEPALIVE_TIME"))
	if err != nil {
		logger.Emit(logging.ERROR, "Failed to parse storage keepalive time: %s", err.Error())
	}

	storageKaTimeout, err := time.ParseDuration(os.Getenv("EXECUTOR_STORAGE_KEEPALIVE_TIMEOUT"))
	if err != nil {
		logger.Emit(logging.ERROR, "Failed to parse storage keepalive timeout: %s", err.Error())
	}

	kv := etcd.NewClient(
		strings.Split(storageEndpoint, ","),
		storageTimeout,
		storageKaTime,
		storageKaTimeout,
	) // Storage client
	logger.Emit(logging.INFO, "Endpoint set to "+"http://"+endpoint+":"+port+"/api/v1/executor")
	c := client.NewClient("http://"+endpoint+":"+port+"/api/v1/executor", logger)
	ex := executor.NewSprintExecutor(nil, nil, c, logger)
	e := events.NewSprintExecutorEventController(ex, make(chan *mesos_v1_executor.Event), logger, kv)
	e.Run()
}

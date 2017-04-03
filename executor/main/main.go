package main

import (
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/executor"
	"mesos-framework-sdk/include/executor"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"os"
	"sprint/executor/events"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	logger := logging.NewDefaultLogger()
	fwId := &mesos_v1.FrameworkID{Value: proto.String(os.Getenv("MESOS_FRAMEWORK_ID"))} // Set implicitly by the Mesos agent.
	execId := &mesos_v1.ExecutorID{Value: proto.String(os.Getenv("MESOS_EXECUTOR_ID"))} // Set implicitly by the Mesos agent.
	endpoint := os.Getenv("EXECUTOR_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost"
	}

	port := os.Getenv("EXECUTOR_ENDPOINT_PORT")
	if port == "" {
		port = "5050"
	}

	logger.Emit(logging.INFO, "Endpoint set to "+"http://"+endpoint+":"+port+"/api/v1/executor")
	c := client.NewClient("http://"+endpoint+":"+port+"/api/v1/executor", logger)
	ex := executor.NewDefaultExecutor(fwId, execId, c, logger)
	e := events.NewSprintExecutorEventController(ex, make(chan *mesos_v1_executor.Event), logger)
	e.Run()
}

package main

import (
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/executor"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/utils"
	"os"
	"sprint/executor/events"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	logger := logging.NewDefaultLogger()

	// Always implicitly set and provided by the Mesos agent.
	fwId := &mesos_v1.FrameworkID{Value: utils.ProtoString(os.Getenv("MESOS_FRAMEWORK_ID"))}
	execId := &mesos_v1.ExecutorID{Value: utils.ProtoString(os.Getenv("MESOS_EXECUTOR_ID"))}
	protocol := os.Getenv("PROTOCOL")
	endpoint := protocol + "://" + os.Getenv("MESOS_AGENT_ENDPOINT") + "/api/v1/executor"

	c := client.NewClient(endpoint, logger)
	ex := executor.NewDefaultExecutor(fwId, execId, c, logger)
	e := events.NewSprintExecutorEventController(ex, logger)
	e.Run()
}

package main

import (
	"mesos-framework-sdk/client"
	exec "mesos-framework-sdk/executor"
	"mesos-framework-sdk/logging"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	logger := logging.NewDefaultLogger()
	c := client.NewClient("http://localhost:5050/api/v1/executor", logger)
	e := exec.NewDefaultExecutor(c, logger)
	e.Run()
}

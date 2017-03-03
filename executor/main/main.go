package main

import (
	//	"flag"
	"mesos-framework-sdk/client"
	exec "mesos-framework-sdk/executor"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	//cfg := new(executor.ExecutorConfiguration).Initialize()
	//flag.Parse()
	c := client.NewClient("http://localhost:5050/api/v1/executor")
	e := exec.NewDefaultExecutor(c)
	e.Run()

}

package main

import (
	"mesos-framework-sdk/client"
	exec "mesos-framework-sdk/executor"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	//cfg := new(exec.ExecutorConfiguration).Initialize()
	//flag.Parse()
	c := client.NewClient("http://68.128.154.89:5050/api/v1/executor")
	e := exec.NewDefaultExecutor(c)
	e.Run()

}

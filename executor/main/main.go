package main

import (
	"flag"
	"sprint/executor"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	cfg := new(executor.ExecutorConfiguration).Initialize()

	flag.Parse()

	executor.NewExecutor(cfg).Run()
}

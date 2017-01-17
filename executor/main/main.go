package main

import (
	"flag"
	"sprint/executor"
)

/*
Main function will wire up all other dependencies for the executor and setup top-level configuration.
*/
func main() {
	// Gather flags for any over-ridden configuration variables.
	flags := flag.NewFlagSet("executor", flag.ExitOnError)

	cfg := new(executor.ExecutorConfiguration).Initialize(flags)
	executor.NewExecutor(cfg).Run()
}

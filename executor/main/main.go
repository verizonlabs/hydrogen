package main

import (
	"flag"
	"log"
	"os"
	"sprint/executor"
)

/*
Main function will wire up all other dependencies for the executor and setup top-level configuration.
*/
func main() {
	// Gather flags for any over-ridden configuration variables.
	flags := flag.NewFlagSet("executor", flag.ExitOnError)

	cfg := new(executor.Configuration).Initialize(flags)

	log.Printf("configuration loaded: %+v", cfg)

	executor.Run(cfg)

	os.Exit(0)
}

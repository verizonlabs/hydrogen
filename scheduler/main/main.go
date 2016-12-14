package main

import (
	"flag"
	"github.com/verizonlabs/sprint/scheduler"
	"log"
	"os"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	fs := flag.NewFlagSet("scheduler", flag.ExitOnError)

	config := new(scheduler.SprintConfiguration)
	config.Initialize(fs)

	fs.Parse(os.Args[1:])

	shutdown := make(chan struct{})
	defer close(shutdown)

	sched := scheduler.NewScheduler(config, shutdown)
	controller := scheduler.NewController(sched, shutdown)
	handlers := scheduler.NewHandlers(sched)

	log.Println("Starting framework scheduler")

	err := sched.Run(controller.SchedulerCtrl(), controller.BuildConfig(
		controller.BuildContext(),
		sched.Caller(),
		shutdown,
		handlers,
	))

	if err != nil {
		log.Fatal(err)
	}
}

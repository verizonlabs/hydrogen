package main

import (
	"flag"
	"log"
	"os"
	"sprint/scheduler"
	"sprint/scheduler/server"
	"strconv"
)

// Serves the executor over HTTP or HTTPS.
func startExecutorServer(config *scheduler.SprintConfiguration) {
	log.Println("New server at " + config.ExecutorSrvPath() + ":" + strconv.Itoa(config.ExecutorSrvPort()))

	if config.ExecutorSrvCert() != "" && config.ExecutorSrvKey() != "" {
		server.NewExecutorServer().ServeExecutorTLS(
			config.ExecutorSrvPath(),
			config.ExecutorSrvPort(),
			config.ExecutorSrvCert(),
			config.ExecutorSrvKey(),
		)
	} else {
		server.NewExecutorServer().ServeExecutor(
			config.ExecutorSrvPath(),
			config.ExecutorSrvPort(),
		)
	}
}

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	fs := flag.NewFlagSet("scheduler", flag.ExitOnError)

	config := new(scheduler.SprintConfiguration).Initialize(fs)

	fs.Parse(os.Args[1:])

	log.Println("Making a new server....")
	go startExecutorServer(config)

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

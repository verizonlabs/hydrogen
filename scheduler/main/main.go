package main

import (
	"flag"
	"log"
	"sprint/scheduler"
	"sprint/scheduler/server"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	schedulerConfig := new(scheduler.SprintConfiguration).Initialize()
	executorSrvConfig := new(server.ServerConfiguration).Initialize()

	flag.Parse()

	schedulerConfig.SetExecutorSrvCfg(executorSrvConfig)

	go server.NewExecutorServer(executorSrvConfig).Serve()

	shutdown := make(chan struct{})
	defer close(shutdown)

	sched := scheduler.NewScheduler(schedulerConfig, shutdown)
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

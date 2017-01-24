package main

import (
	"flag"
	"log"
	"sprint/scheduler"
	"sprint/scheduler/server"
	"sprint/scheduler/server/api"
	"sprint/scheduler/server/file"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	srvConfig := new(server.ServerConfiguration).Initialize()
	schedulerConfig := new(scheduler.SchedulerConfiguration).Initialize().SetExecutorSrvCfg(srvConfig)

	executorSrv := file.NewExecutorServer(srvConfig)
	apiSrv := api.NewApiServer(srvConfig)

	// Parse here to catch flags defined in structures above.
	flag.Parse()

	shutdown := make(chan struct{})
	defer close(shutdown)

	sched := scheduler.NewScheduler(schedulerConfig, shutdown)
	controller := scheduler.NewController(sched, shutdown)
	handlers := scheduler.NewHandlers(sched)

	log.Println("Starting executor server...")
	go executorSrv.Serve()
	log.Println("Starting API server...")
	go apiSrv.RunAPI()

	log.Println("Starting framework scheduler...")
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

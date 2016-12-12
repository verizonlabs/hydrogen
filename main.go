package main

import (
	"flag"
	"github.com/verizonlabs/sprint/scheduler"
	"os"
)

func main() {
	component := flag.String("component", "scheduler", "The component to run [scheduler|executor]")
	flag.Parse()

	switch *component {
	case "scheduler":
		fs := flag.NewFlagSet("scheduler", flag.ExitOnError)

		config := new(scheduler.Configuration)
		config.Initialize(fs)

		fs.Parse(os.Args[2:])

		shutdown := make(chan struct{})

		sched := scheduler.NewScheduler(config, shutdown)
		controller := scheduler.NewController(shutdown)
		sched.Run(controller.GetSchedulerCtrl(), controller.BuildConfig(
			controller.BuildContext(),
			controller.BuildFrameworkInfo(config),
			sched.GetCaller(),
			shutdown,
		))
	case "executor":
		//TODO start executor here
	}
}

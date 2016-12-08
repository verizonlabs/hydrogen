package scheduler


import (
	sched "github.com/verizonlabs/mesos-go/scheduler"
	"github.com/verizonlabs/mesos-go/mesosproto"
)

func NewSchedulerDriver(scheduler sched.Scheduler, frameworkInfo *mesosproto.FrameworkInfo, master string) (*sched.MesosSchedulerDriver, error){
	driverConfig := sched.DriverConfig{
		Scheduler: scheduler,
		Framework: frameworkInfo,
		Master: master,
	}

	return sched.NewMesosSchedulerDriver(driverConfig)
}
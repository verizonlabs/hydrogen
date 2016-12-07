package scheduler

import (
	mesos "github.com/verizonlabs/mesos-go/mesosproto"
	sched "github.com/verizonlabs/mesos-go/scheduler"
)

type scheduler struct {
	executor *mesos.ExecutorInfo
}

func NewScheduler(executor *mesos.ExecutorInfo) *scheduler {
	return &scheduler{
		executor: executor,
	}
}

func (s *scheduler) Registered(driver sched.SchedulerDriver, frameworkId *mesos.FrameworkID, master *mesos.MasterInfo) {
	//Registered
}

func (s *scheduler) Reregistered(driver sched.SchedulerDriver, master *mesos.MasterInfo) {
	//Re-registered
}

func (s *scheduler) Disconnected(driver sched.SchedulerDriver) {
	//Disconnected
}

func (s *scheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	//Resource offers
}

func (s *scheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	//Task status updates
}

func (s *scheduler) OfferRescinded(driver sched.SchedulerDriver, offerId *mesos.OfferID) {
	//Rescind offer
}

func (s *scheduler) FrameworkMessage(driver sched.SchedulerDriver, executorId *mesos.ExecutorID, slaveId *mesos.SlaveID, msg string) {
	//Messages
}

func (s *scheduler) SlaveLost(driver sched.SchedulerDriver, slaveId *mesos.SlaveID) {
	//Lost slave
}

func (s *scheduler) ExecutorLost(driver sched.SchedulerDriver, executorId *mesos.ExecutorID, slaveId *mesos.SlaveID, code int) {
	//Lost executor
}

func (s *scheduler) Error(driver sched.SchedulerDriver, err string) {
	//Scheduler got an error
}

package scheduler

import (
	mesos "github.com/verizonlabs/mesos-go/mesosproto"
	"github.com/verizonlabs/mesos-go/scheduler"
)

type Scheduler struct {
	executor *mesos.ExecutorInfo
}

func NewScheduler(executor *mesos.ExecutorInfo) *Scheduler {
	return &Scheduler{
		executor: executor,
	}
}

func (s *Scheduler) Registered(driver scheduler.SchedulerDriver, frameworkId *mesos.FrameworkID, master *mesos.MasterInfo) {
	//Registered
}

func (s *Scheduler) Reregistered(driver scheduler.SchedulerDriver, master *mesos.MasterInfo) {
	//Re-registered
}

func (s *Scheduler) Disconnected(driver scheduler.SchedulerDriver) {
	//Disconnected
}

func (s *Scheduler) ResourceOffers(driver scheduler.SchedulerDriver, offers []*mesos.Offer) {
	//Resource offers
}

func (s *Scheduler) StatusUpdate(driver scheduler.SchedulerDriver, status *mesos.TaskStatus) {
	//Task status updates
}

func (s *Scheduler) OfferRescinded(driver scheduler.SchedulerDriver, offerId *mesos.OfferID) {
	//Rescind offer
}

func (s *Scheduler) FrameworkMessage(driver scheduler.SchedulerDriver, executorId *mesos.ExecutorID, slaveId *mesos.SlaveID, msg string) {
	//Messages
}

func (s *Scheduler) SlaveLost(driver scheduler.SchedulerDriver, slaveId *mesos.SlaveID) {
	//Lost slave
}

func (s *Scheduler) ExecutorLost(driver scheduler.SchedulerDriver, executorId *mesos.ExecutorID, slaveId *mesos.SlaveID, code int) {
	//Lost executor
}

func (s *Scheduler) Error(driver scheduler.SchedulerDriver, err string) {
	//Scheduler got an error
}

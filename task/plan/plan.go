package plan

import (
	schedulerConfig "github.com/verizonlabs/hydrogen/scheduler"
	"github.com/verizonlabs/hydrogen/scheduler/ha"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/persistence"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/scheduler"
)

const (
	Idle      PlanType = 0
	Launch    PlanType = 1
	Update    PlanType = 2
	Kill      PlanType = 3
	Reconcile PlanType = 4
	Subscribe PlanType = 5

	QUEUED   PlanState = 0
	RUNNING  PlanState = 1
	FINISHED PlanState = 2
	ERROR    PlanState = 3
	RETRY    PlanState = 4
)

type (
	PlanType  uint8
	PlanState uint8

	// Any plan needs to be able to control the lifecycle of a task.
	Plan interface {
		Execute() error    // Execute the current plan.
		Update() error     // Update the current plan status.
		Status() PlanState // Current status of the plan.
		Type() PlanType    // Tells you the plan type.
	}
)

func NewPlanFactory(
	taskMgr manager.TaskManager,
	resourceMgr resource.ResourceManager,
	scheduler scheduler.Scheduler,
	storage persistence.Storage,
	ha ha.HighAvailablity,
	configuration schedulerConfig.Configuration,
	logger logging.Logger,
	events chan *mesos_v1_scheduler.Event) func([]*manager.Task, PlanType) Plan {

	return func(tasks []*manager.Task, planType PlanType) Plan {
		switch planType {
		case Idle:
			return NewIdlePlan()
		case Launch:
			return NewLaunchPlan(tasks, taskMgr, resourceMgr, scheduler, storage)
		case Update:
			return NewUpdatePlan(tasks, taskMgr, resourceMgr, scheduler, storage)
		case Kill:
			return NewKillPlan(tasks, taskMgr, resourceMgr, scheduler, storage)
		case Reconcile:
			return NewReconcilePlan(tasks, taskMgr, resourceMgr, scheduler, storage, configuration, logger)
		case Subscribe:
			return NewSubscribePlan(tasks, taskMgr, resourceMgr, scheduler, storage, ha, logger, configuration, events)
		}
	}
}

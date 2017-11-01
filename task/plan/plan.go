package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/scheduler"
)

const (
	Idle      PlanType = 0
	Launch    PlanType = 1
	Update    PlanType = 2
	Kill      PlanType = 3
	Reconcile PlanType = 4

	IDLE     PlanState = 0
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
	scheduler scheduler.Scheduler) func([]*manager.Task, PlanType) Plan {

	return func(tasks []*manager.Task, planType PlanType) Plan {
		switch planType {
		case Idle:
			return NewIdlePlan()
		case Launch:
			return NewLaunchPlan(tasks, taskMgr, resourceMgr, scheduler)
		case Update:
			return NewUpdatePlan(tasks, taskMgr, resourceMgr, scheduler)
		case Kill:
			return NewKillPlan(tasks, taskMgr, resourceMgr, scheduler)
		case Reconcile:
			return NewReconcilePlan(tasks, taskMgr, resourceMgr, scheduler)
		}
	}
}

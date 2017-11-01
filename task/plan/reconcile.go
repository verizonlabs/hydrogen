package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/scheduler"
)

type ReconcilePlan struct {
	Plan
}

func NewReconcilePlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	i scheduler.Scheduler) Plan {
	return ReconcilePlan{}
}

func (r *ReconcilePlan) Execute() error    {} // Execute the current plan.
func (r *ReconcilePlan) Update() error     {} // Update the current plan status.
func (r *ReconcilePlan) Status() PlanState {} // Current status of the plan.
func (r *ReconcilePlan) Type() PlanType    {} // Tells you the plan type.

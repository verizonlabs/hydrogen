package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/scheduler"
)

type UpdatePlan struct {
	Plan
}

func NewUpdatePlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	i scheduler.Scheduler) Plan {
	return UpdatePlan{}
}

func (u *UpdatePlan) Execute() error    {} // Execute the current plan.
func (u *UpdatePlan) Update() error     {} // Update the current plan status.
func (u *UpdatePlan) Status() PlanState {} // Current status of the plan.
func (u *UpdatePlan) Type() PlanType    {} // Tells you the plan type.

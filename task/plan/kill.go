package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/scheduler"
)

type KillPlan struct {
	Plan
}

func NewKillPlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	i scheduler.Scheduler) Plan {

	return KillPlan{}
}

func (k *KillPlan) Execute() error    {} // Execute the current plan.
func (k *KillPlan) Update() error     {} // Update the current plan status.
func (k *KillPlan) Status() PlanState {} // Current status of the plan.
func (k *KillPlan) Type() PlanType    {} // Tells you the plan type.

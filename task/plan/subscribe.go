package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/scheduler"
)

type SubscribePlan struct {
	Plan
	tasks           []*manager.Task
	taskManager     manager.TaskManager
	resourceManager resource.ResourceManager
	scheduler       scheduler.Scheduler
}

func NewSubscribePlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	scheduler scheduler.Scheduler) Plan {
	return SubscribePlan{
		tasks:           tasks,
		taskManager:     taskManager,
		resourceManager: resourceManager,
		scheduler:       scheduler,
	}
}

func (k *SubscribePlan) Execute() error    {} // Execute the current plan.
func (k *SubscribePlan) Update() error     {} // Update the current plan status.
func (k *SubscribePlan) Status() PlanState {} // Current status of the plan.
func (k *SubscribePlan) Type() PlanType    {} // Tells you the plan type.

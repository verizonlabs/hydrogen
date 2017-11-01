package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
)

type (
	LaunchPlan struct {
		Plan
		task            []*manager.Task
		taskManager     manager.TaskManager
		resourceManager resource.ResourceManager
		scheduler       scheduler.Scheduler
	}
)

func NewLaunchPlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	scheduler scheduler.Scheduler) Plan {

	return &LaunchPlan{
		task:            tasks,
		taskManager:     taskManager,
		resourceManager: resourceManager,
		scheduler:       scheduler,
	}
}

func (d *LaunchPlan) Execute() error {
	// At this point we can be assured our task is valid and has been validated by the parser.
	// We can safely add it to the task manager.
	err := d.taskManager.Add(d.task...)
	if err != nil {
		return err
	}

	d.scheduler.Revive()
	return nil
}

// Here we can update the status of the plan.
func (d *LaunchPlan) Update() error {
	return nil
}

func (d *LaunchPlan) Status() PlanState {} // Current status of the plan.
func (d *LaunchPlan) Type() PlanType    {} // Tells you the plan type.

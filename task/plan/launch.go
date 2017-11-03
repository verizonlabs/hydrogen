package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/persistence"
)

type (
	LaunchPlan struct {
		Plan
		task            []*manager.Task
		taskManager     manager.TaskManager
		resourceManager resource.ResourceManager
		scheduler       scheduler.Scheduler
		state           PlanState
	}
)

func NewLaunchPlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	scheduler scheduler.Scheduler,
	storage persistence.Storage) Plan {

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
	d.state = RUNNING
	err := d.taskManager.Add(d.task...)
	if err != nil {
		return err
	}

	d.scheduler.Revive()

	// We want to wait to see if the task is launched by the framework or not before setting complete.
	// TODO (tim): How do we get the event updates routed to each plan?
	// We could wait on a channel for an update.


	return nil
}

// Here we can update the status of the plan.
func (d *LaunchPlan) Update() error {
	return nil
}

func (d *LaunchPlan) Status() PlanState {

} // Current status of the plan.

func (d *LaunchPlan) Type() PlanType {

} // Tells you the plan type.

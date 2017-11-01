package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
)

type (
	DeploymentPlan struct {
		task            []*manager.Task
		taskManager     manager.TaskManager
		resourceManager resource.ResourceManager
		scheduler       scheduler.Scheduler
	}
)

// Options for deployment are held within the task itself.
func NewDeploymentPlan(task []*manager.Task,
	taskMgr manager.TaskManager,
	resourceMgr resource.ResourceManager,
	scheduler scheduler.Scheduler) *DeploymentPlan {

	return &DeploymentPlan{
		task:            task,
		taskManager:     taskMgr,
		resourceManager: resourceMgr,
		scheduler:       scheduler,
	}
}

// TODO (tim): How do we handle multiple tasks or group execution plans?
// Perhaps we have a separate multiple deployment plan.
func (d *DeploymentPlan) Execute() error {
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
func (d *DeploymentPlan) Update() error {
	return nil
}

package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/persistence"
	"mesos-framework-sdk/scheduler"
	"errors"
)

type KillPlan struct {
	Plan
}

func NewKillPlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	scheduler scheduler.Scheduler,
	storage persistence.Storage) Plan {

	return KillPlan{}
}

func (k *KillPlan) Execute() error    {
	for _, iterTask := range appJSON {
		// Make sure we have a name to look up
		if iterTask.Name == nil {
			return "", errors.New("Task name is nil")
		}

		// Look up task in task manager
		tsk, err := m.taskManager.Get(iterTask.Name)
		if err != nil {
			return "", err
		}

		err = m.taskManager.Delete(tsk)
		if err != nil {
			return "", err
		}

		// This stops on first error, do we want to do best-effort clean up anyways (i.e. continue after single error?)
		if tsk.State != t.UNKNOWN {
			_, err := m.scheduler.Kill(tsk.Info.GetTaskId(), tsk.Info.GetAgentId())
			if err != nil {
				return "", err
			}
		}
	}
	// TODO (tim): Return multiple names.
} // Execute the current plan.
func (k *KillPlan) Update() error     {} // Update the current plan status.
func (k *KillPlan) Status() PlanState {} // Current status of the plan.
func (k *KillPlan) Type() PlanType    {} // Tells you the plan type.

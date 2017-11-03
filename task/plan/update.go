package plan

import (
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/persistence"
	"mesos-framework-sdk/scheduler"
)

type UpdatePlan struct {
	Plan
}

func NewUpdatePlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	scheduler scheduler.Scheduler,
	storage persistence.Storage) Plan {
	return UpdatePlan{}
}

func (u *UpdatePlan) Execute() error {
	/*
			Current behavior is to treat each task seperately.
			If we fail on the 3rd task out of 5 to update,
			3 are in the old version, 2 are on the new version.
			Is this desirable? Or should we roll back upon any
			failure within our defer function? Adjustible?

		for _, iterTask := range appJSON {
			taskToKill, err := m.taskManager.Get(&iterTask.Name)
			if err != nil {
				return nil, err
			}
			// If the old task exists, parse the new one.

			// Kill the old one.
			resp, err := m.scheduler.Kill(taskToKill.Info.GetTaskId(), taskToKill.Info.GetAgentId())
			fmt.Fprintf(os.Stdout, "resp code %v, err %v\n", resp.StatusCode, err)
			if resp.StatusCode == 202 {
				err := m.taskManager.Delete(mesosTask...)
				if err != nil {
					fmt.Fprintf(os.Stdout, "Broke on del, %v\n", err.Error())
				}
				err = m.taskManager.Add(mesosTask...)
				fmt.Fprintf(os.Stdout, "????? %v\n\n", err)
				if err != nil {
					fmt.Fprintf(os.Stdout, "Broke on add, %v\n", err.Error())
				}
				if !success {
					success = true
				}
				tp, err := m.taskManager.Get(mesosTask[0].Info.Name)
				fmt.Fprintf(os.Stdout, "GET :  name %v err %v\n", *tp.Info.Name, err)
			}
		}
		//TODO(tim): Return multiple names.
	*/
}                                       // Execute the current plan.
func (u *UpdatePlan) Update() error     {} // Update the current plan status.
func (u *UpdatePlan) Status() PlanState {} // Current status of the plan.
func (u *UpdatePlan) Type() PlanType    {} // Tells you the plan type.

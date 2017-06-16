package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Update(t *testing.T) {
	ctrl := workingEventControllerFactory()
	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
			State:  state,
		}})
	}
}

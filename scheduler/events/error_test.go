package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Error(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Error(&mesos_v1_scheduler.Event_Error{
		Message: utils.ProtoString("message"),
	})
}

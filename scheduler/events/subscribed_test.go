package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Subscribe(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Subscribed(&mesos_v1_scheduler.Event_Subscribed{FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("id")}})
}

func TestSprintEventController_Subscribed(t *testing.T) {
	ctrl := workingEventControllerFactory()
	go ctrl.Run()
	c <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}
}

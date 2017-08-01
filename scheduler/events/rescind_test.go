package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Rescind(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.Run()

	// We need to subscribe first before receiving any other events.
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}

	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: &mesos_v1_scheduler.Event_Rescind{},
	}
}

func TestSprintEventController_RescindWithNil(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.Run()

	// We need to subscribe first before receiving any other events.
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}

	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: &mesos_v1_scheduler.Event_Rescind{OfferId: nil},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: nil,
	}
}

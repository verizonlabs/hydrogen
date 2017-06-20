package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"testing"
)

func TestSprintEventController_Rescind(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.Run()
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: &mesos_v1_scheduler.Event_Rescind{},
	}
}

func TestSprintEventController_RescindWithNil(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.Run()
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: &mesos_v1_scheduler.Event_Rescind{OfferId: nil},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: nil,
	}
}

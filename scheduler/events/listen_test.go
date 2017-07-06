package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Listen(t *testing.T) {
	ctrl := workingEventController()

	go ctrl.Run()

	ctrl.events <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}

	// Test all event messages.
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:  mesos_v1_scheduler.Event_ERROR.Enum(),
		Error: &mesos_v1_scheduler.Event_Error{},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_FAILURE.Enum(),
		Failure: &mesos_v1_scheduler.Event_Failure{},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:          mesos_v1_scheduler.Event_INVERSE_OFFERS.Enum(),
		InverseOffers: &mesos_v1_scheduler.Event_InverseOffers{},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_MESSAGE.Enum(),
		Message: &mesos_v1_scheduler.Event_Message{},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:   mesos_v1_scheduler.Event_OFFERS.Enum(),
		Offers: &mesos_v1_scheduler.Event_Offers{},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: &mesos_v1_scheduler.Event_Rescind{},
	}
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type:   mesos_v1_scheduler.Event_UPDATE.Enum(),
		Update: &mesos_v1_scheduler.Event_Update{},
	}
}

package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_InverseOffer(t *testing.T) {
	ctrl := workingEventController()
	ctrl.InverseOffer(&mesos_v1_scheduler.Event_InverseOffers{
		InverseOffers: []*mesos_v1.InverseOffer{
			{
				Id:             &mesos_v1.OfferID{Value: utils.ProtoString("id")},
				FrameworkId:    &mesos_v1.FrameworkID{Value: utils.ProtoString("id")},
				Unavailability: &mesos_v1.Unavailability{Start: &mesos_v1.TimeInfo{Nanoseconds: utils.ProtoInt64(0)}},
			},
		},
	})
}

func TestSprintEventController_InverseOfferWithNilOffer(t *testing.T) {
	ctrl := workingEventController()
	ctrl.InverseOffer(&mesos_v1_scheduler.Event_InverseOffers{InverseOffers: nil})
	ctrl.InverseOffer(nil)
}

package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_RescindInverseOffer(t *testing.T) {
	ctrl := workingEventController()
	ctrl.RescindInverseOffer(&mesos_v1_scheduler.Event_RescindInverseOffer{
		InverseOfferId: &mesos_v1.OfferID{Value: utils.ProtoString("id")},
	})
}

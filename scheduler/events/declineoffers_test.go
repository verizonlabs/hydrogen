package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"testing"
)

func TestSprintEventController_declineOffers(t *testing.T) {
	ctrl := workingEventController()
	ctrl.declineOffers(make([]*mesos_v1.Offer, 0), 0.0)
}

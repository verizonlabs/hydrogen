package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Offers(t *testing.T) {
	ctrl := workingEventControllerFactory()

	// Test empty offers.
	offers := []*mesos_v1.Offer{}
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

	// Test a single offer
	offers = []*mesos_v1.Offer{}
	resources := []*mesos_v1.Resource{}
	resources = append(resources, &mesos_v1.Resource{
		Name: utils.ProtoString("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: utils.ProtoFloat64(10.0),
		},
	})
	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: utils.ProtoString("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: utils.ProtoString(utils.UuidAsString())},
		Hostname:    utils.ProtoString("Some host"),
		Resources:   resources,
	})
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

}

func TestSprintEventController_OffersWithQueuedTasks(t *testing.T) {
	ctrl := workingEventControllerFactory()
	// Test empty offers.
	offers := []*mesos_v1.Offer{}
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

	// Test a single offer
	offers = []*mesos_v1.Offer{}
	resources := []*mesos_v1.Resource{}
	resources = append(resources, &mesos_v1.Resource{
		Name: utils.ProtoString("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: utils.ProtoFloat64(10.0),
		},
	})
	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: utils.ProtoString("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: utils.ProtoString(utils.UuidAsString())},
		Hostname:    utils.ProtoString("Some host"),
		Resources:   resources,
	})
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

}

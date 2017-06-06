package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/utils"
)

//
// Decline offers is a private method to organize a list of offers that are to be declined by the
// scheduler.
//
func (s *SprintEventController) declineOffers(offers []*mesos_v1.Offer, refuseSeconds float64) {
	if len(offers) == 0 {
		return
	}

	declineIDs := []*mesos_v1.OfferID{}

	// Decline whatever offers are left over
	for _, id := range offers {
		declineIDs = append(declineIDs, id.GetId())
	}

	s.scheduler.Decline(declineIDs, &mesos_v1.Filters{RefuseSeconds: utils.ProtoFloat64(refuseSeconds)})
}

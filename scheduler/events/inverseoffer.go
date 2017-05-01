package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
)

//
// Inverse offers
//
func (s *SprintEventController) InverseOffer(ioffers *mesos_v1_scheduler.Event_InverseOffers) {
	s.logger.Emit(logging.INFO, "Inverse Offer event recieved: %v", ioffers)
}

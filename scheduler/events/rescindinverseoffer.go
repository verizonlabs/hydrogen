package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
)

//
// Rescind Inverse Offers is a public method that handles the event
// when an inverse offer isn't declined or accepted within the time out period set.
//
func (s *SprintEventController) RescindInverseOffer(rioffers *mesos_v1_scheduler.Event_RescindInverseOffer) {
	s.logger.Emit(logging.INFO, "Rescind Inverse Offer event recieved: %v", rioffers)
}

package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
)

//
// Rescind is a public method that handles a rescind event from the mesos-master.
// Rescind events only occur if an offer isn't declined properly within the offer
// timeout period.
//
func (s *SprintEventController) Rescind(rescindEvent *mesos_v1_scheduler.Event_Rescind) {
	s.logger.Emit(logging.INFO, "Rescind event recieved: %v", *rescindEvent)
}

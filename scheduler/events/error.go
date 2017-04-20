package events

import (
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
)

//
// Error is a public method that handles an error event sent by the mesos master.
// We log errors to the logger, and handle any special cases as needed here.
//
func (s *SprintEventController) Error(err *mesos_v1_scheduler.Event_Error) {
	s.logger.Emit(logging.ERROR, "Error event recieved: %v", err)
}

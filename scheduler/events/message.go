package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
)

//
// Message is a public method that handles a message event from the mesos master.
// We simply print this message to the log, however in the future we may utilize
// the message passing functionality to send custom logging errors.
//
func (s *SprintEventController) Message(msg *mesos_v1_scheduler.Event_Message) {
	s.logger.Emit(logging.INFO, "Message event recieved: %v", *msg)
}

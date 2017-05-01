package events

import (
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
)

//
// Failure is a public method that respond to a Failure event sent by the mesos master.
// We log the failure here with the logger.
//
func (s *SprintEventController) Failure(fail *mesos_v1_scheduler.Event_Failure) {
	s.logger.Emit(logging.ERROR, "Executor %s failed with status %d", fail.GetExecutorId().GetValue(), fail.GetStatus())
}

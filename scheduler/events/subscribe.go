package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	"os"
)

//
// Subscribe is a public method that handles subscription events from the master.
// We handle our subscription by writing the framework ID to storage.
// Then we gather all of our launched non-terminal tasks and reconcile them explicitly.
//
func (s *SprintEventController) Subscribe(subEvent *mesos_v1_scheduler.Event_Subscribed) {
	id := subEvent.GetFrameworkId()
	idVal := id.GetValue()
	s.scheduler.FrameworkInfo().Id = id
	s.logger.Emit(logging.INFO, "Subscribed with an ID of %s", idVal)

	err := s.createFrameworkIdLease(idVal)
	if err != nil {
		s.logger.Emit(logging.ERROR, "Failed to persist leader information: %s", err.Error())
		os.Exit(3)
	}

	// Get all launched non-terminal tasks.
	launched, err := s.taskmanager.GetState(manager.RUNNING)
	if err != nil {
		s.logger.Emit(logging.INFO, "Not reconciling: %s", err.Error())
		return
	}

	// Reconcile after we subscribe in case we resubscribed due to a failure.
	s.scheduler.Reconcile(launched)
}

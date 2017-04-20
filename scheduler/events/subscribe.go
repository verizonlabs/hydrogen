package events

import (
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	"time"
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

	var lease int64
	var err error
	for {
		lease, err = s.kv.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.FrameworkInfo().GetFailoverTimeout()))
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		break
	}

	s.frameworkLease = lease

	// Get all launched non-terminal tasks.
	launched, err := s.taskmanager.GetState(manager.RUNNING)
	if err != nil {
		s.logger.Emit(logging.INFO, "Not reconciling: %s", err.Error())
		return
	}

	// Reconcile after we subscribe in case we resubscribed due to a failure.
	s.scheduler.Reconcile(launched)
}

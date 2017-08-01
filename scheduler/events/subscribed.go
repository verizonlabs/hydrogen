package events

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	"os"
	"mesos-framework-sdk/include/mesos_v1"
)

//
// Subscribe is a public method that handles subscription events from the master.
// We handle our subscription by writing the framework ID to storage.
// Then we gather all of our launched non-terminal tasks and reconcile them explicitly.
//
func (s *SprintEventController) Subscribed(subEvent *mesos_v1_scheduler.Event_Subscribed) {
	id := subEvent.GetFrameworkId()
	idVal := id.GetValue()
	s.scheduler.FrameworkInfo().Id = id
	s.logger.Emit(logging.INFO, "Subscribed with an ID of %s", idVal)

	err := s.createFrameworkIdLease(idVal)
	if err != nil {
		s.logger.Emit(logging.ERROR, "Failed to persist leader information: %s", err.Error())
		os.Exit(3)
	}

	s.scheduler.Revive() // Reset to revive offers regardless if there are tasks or not.
	// We do this to force a check for any tasks that we might have missed during downtime.
	// Reconcile after we subscribe in case we resubscribed due to a failure.

	// Get all launched non-terminal tasks.
	launched, err := s.taskmanager.AllByState(manager.RUNNING)
	if err != nil {
		s.logger.Emit(logging.INFO, "Not reconciling: %s", err.Error())
		return
	}

	toReconcile :=  []*mesos_v1.TaskInfo{}
	for _, t := range launched {
		toReconcile = append(toReconcile, t.Info)
	}
	s.scheduler.Reconcile(toReconcile)
}

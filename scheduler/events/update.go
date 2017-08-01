package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
)

//
// Update is a public method that handles an update event from the mesos master.
// Depending on the update event, we handle the event as is appropriate.
//
func (s *SprintEventController) Update(updateEvent *mesos_v1_scheduler.Event_Update) {
	status := updateEvent.GetStatus()
	agentID := status.GetAgentId()
	taskID := status.GetTaskId()

	// Always acknowledge that we've received the message from Mesos.
	defer func() {
		_, err := s.scheduler.Acknowledge(agentID, taskID, status.GetUuid())
		if err != nil {
			s.logger.Emit(
				logging.ERROR,
				"Failed to acknowledge event from agent %s for task ID %s: %s",
				agentID.GetValue(),
				taskID.GetValue(),
				err.Error(),
			)
		}
	}()

	task, err := s.taskmanager.GetById(taskID)
	if err != nil {
		// The event is from a task that has been deleted from the task manager,
		// ignore updates.
		// NOTE (tim): Do we want to keep deleted task history for a certain amount of time
		// before it's deleted? We would record status updates after it's killed here.
		// ACK update, return.
		return
	}

	state := status.GetState()
	message := status.GetMessage()
	taskIdVal := taskID.GetValue()
	agentIdVal := agentID.GetValue()

	// Update the state of the task.
	task.State = state

	if err != nil {
		s.logger.Emit(logging.ERROR, "Failed to update task %s: %s", taskIdVal, err.Error())
		return
	}

	switch state {
	case mesos_v1.TaskState_TASK_FAILED:
		s.logger.Emit(logging.ERROR, "Task %s failed: %s", taskIdVal, message)
		s.reschedule(task)
	case mesos_v1.TaskState_TASK_STAGING:
		// NOP, keep task set to "launched".
		s.logger.Emit(logging.INFO, "Task %s is staging: %s", taskIdVal, message)
	case mesos_v1.TaskState_TASK_DROPPED:

		// Transient error, we should retry launching. Taskinfo is fine.
		s.logger.Emit(logging.INFO, "Task %s dropped: %s", taskIdVal, message)
		s.reschedule(task)
	case mesos_v1.TaskState_TASK_ERROR:
		s.logger.Emit(logging.ERROR, "Error with task %s: %s", taskIdVal, message)
		s.reschedule(task)
	case mesos_v1.TaskState_TASK_FINISHED:
		s.logger.Emit(
			logging.INFO,
			"Task %s finished on agent %s: %s",
			taskIdVal,
			agentIdVal,
			message,
		)
		s.taskmanager.Delete(task)
	case mesos_v1.TaskState_TASK_GONE:
		// Agent is dead and task is lost.
		s.logger.Emit(logging.ERROR, "Task %s is gone: %s", taskIdVal, message)

		s.reschedule(task)
	case mesos_v1.TaskState_TASK_GONE_BY_OPERATOR:
		// Agent might be dead, master is unsure. Will return to RUNNING state possibly or die.
		s.logger.Emit(logging.ERROR, "Task %s gone by operator: %s", taskIdVal, message)
	case mesos_v1.TaskState_TASK_KILLED:
		// Task was killed.
		s.logger.Emit(
			logging.INFO,
			"Task %s on agent %s was killed",
			taskIdVal,
			agentIdVal,
		)
		s.taskmanager.Delete(task)
	case mesos_v1.TaskState_TASK_KILLING:

		// Task is in the process of catching a SIGNAL and shutting down.
		s.logger.Emit(logging.INFO, "Killing task %s: %s", taskIdVal, message)
	case mesos_v1.TaskState_TASK_LOST:
		// Task is unknown to the master and lost. Should reschedule.
		s.logger.Emit(logging.ALARM, "Task %s was lost", taskIdVal)

		s.reschedule(task)
	case mesos_v1.TaskState_TASK_RUNNING:
		s.logger.Emit(
			logging.INFO,
			"Task %s is running on agent %s",
			taskIdVal,
			agentIdVal,
		)
	case mesos_v1.TaskState_TASK_STARTING:

		// Task is still starting up. NOOP
		s.logger.Emit(logging.INFO, "Task %s is starting: %s", taskIdVal, message)
	case mesos_v1.TaskState_TASK_UNKNOWN:

		// Task is unknown to the master. Should ignore.
		s.logger.Emit(logging.ALARM, "Task %s is unknown: %s", taskIdVal, message)
	case mesos_v1.TaskState_TASK_UNREACHABLE:

		// Agent lost contact with master, could be a network error. No guarantee the task is still running.
		// Should we reschedule after waiting a certain period of time?
		s.logger.Emit(logging.INFO, "Task %s is unreachable: %s", taskIdVal, message)
	}
}

// Sets a task to be rescheduled.
// Rescheduling can be done when there are various failures such as network errors.
func (s *SprintEventController) reschedule(task *manager.Task) {
	// We need to check for nil tasks in order for testing to work with mock types.

	task.Reschedule()
	s.taskmanager.Update(task)
	s.Scheduler().Revive()

}

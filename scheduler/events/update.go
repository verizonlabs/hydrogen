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
	agentId := status.GetAgentId()
	taskId := status.GetTaskId()
	task, err := s.taskmanager.GetById(taskId)
	if err != nil {
		// The event is from a task that has been deleted from the task manager,
		// ignore updates.
		// NOTE (tim): Do we want to keep deleted task history for a certain amount of time
		// before it's deleted? We would record status updates after it's killed here.
		// ACK update, return.
		s.scheduler.Acknowledge(agentId, taskId, status.GetUuid())
		return
	}

	state := status.GetState()
	message := status.GetMessage()
	taskIdVal := taskId.GetValue()
	agentIdVal := agentId.GetValue()

	s.taskmanager.Set(state, task)

	switch state {
	case mesos_v1.TaskState_TASK_FAILED:
		if s.taskmanager.IsInGroup(task) {
			s.taskmanager.Unlink(task.GetName(), agentId)
		}
		s.logger.Emit(logging.ERROR, "Task %s failed: %s", taskIdVal, message)
		s.reschedule(task)
	case mesos_v1.TaskState_TASK_STAGING:
		// NOP, keep task set to "launched".
		s.logger.Emit(logging.INFO, "Task %s is staging: %s", taskIdVal, message)
	case mesos_v1.TaskState_TASK_DROPPED:
		if s.taskmanager.IsInGroup(task) {
			s.taskmanager.Unlink(task.GetName(), agentId)
		}
		// Transient error, we should retry launching. Taskinfo is fine.
		s.logger.Emit(logging.INFO, "Task %s dropped: %s", taskIdVal, message)
		s.reschedule(task)
	case mesos_v1.TaskState_TASK_ERROR:
		if s.taskmanager.IsInGroup(task) {
			s.taskmanager.Unlink(task.GetName(), agentId)
		}
		s.logger.Emit(logging.ERROR, "Error with task %s: %s", taskIdVal, message)
		s.reschedule(task)
	case mesos_v1.TaskState_TASK_FINISHED:
		if s.taskmanager.IsInGroup(task) {
			s.taskmanager.Unlink(task.GetName(), agentId)
			s.taskmanager.SetSize(task.GetName(), -1)
			// We only delete the entire group if the size is 0.
			s.taskmanager.DeleteGroup(task.GetName())
		}
		s.logger.Emit(
			logging.INFO,
			"Task %s finished on agent %s: %s",
			taskIdVal,
			agentIdVal,
			message,
		)
		s.taskmanager.Delete(task)
	case mesos_v1.TaskState_TASK_GONE:
		if s.taskmanager.IsInGroup(task) {
			s.taskmanager.Unlink(task.GetName(), agentId)
		}
		// Agent is dead and task is lost.
		s.logger.Emit(logging.ERROR, "Task %s is gone: %s", taskIdVal, message)
		s.reschedule(task)
	case mesos_v1.TaskState_TASK_GONE_BY_OPERATOR:
		// Agent might be dead, master is unsure. Will return to RUNNING state possibly or die.
		s.logger.Emit(logging.ERROR, "Task %s gone by operator: %s", taskIdVal, message)
	case mesos_v1.TaskState_TASK_KILLED:
		if s.taskmanager.IsInGroup(task) {
			s.taskmanager.Unlink(task.GetName(), agentId)
			s.taskmanager.SetSize(task.GetName(), -1)
			s.taskmanager.DeleteGroup(task.GetName())
		}
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
		// A task can be lost if it never got to the master.
		if s.taskmanager.IsInGroup(task) {
			s.taskmanager.Unlink(task.GetName(), agentId)
		}
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
	default:
	}

	s.scheduler.Acknowledge(agentId, taskId, status.GetUuid())
}

// Sets a task to be rescheduled.
// Rescheduling can be done when there are various failures such as network errors.
func (s *SprintEventController) reschedule(task *mesos_v1.TaskInfo) {
	policy := s.taskmanager.CheckPolicy(task)
	retryFunc := func() error {

		// Check if the task has been deleted while waiting for a retry.
		t, err := s.taskmanager.Get(task.Name)
		if err != nil {
			return err
		}
		s.taskmanager.Set(manager.UNKNOWN, t)
		s.Scheduler().Revive()

		return nil
	}
	err := s.taskmanager.RunPolicy(policy, retryFunc)
	if err != nil {
		s.logger.Emit(logging.ERROR, "Failed to run policy: %s", err.Error())
	}
}

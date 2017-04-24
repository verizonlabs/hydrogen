package events

import (
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
)

//
// Update is a public method that handles an update event from the mesos master.
// Depending on the update event, we handle the event as is appropriate.
//
func (s *SprintEventController) Update(updateEvent *mesos_v1_scheduler.Event_Update) {
	status := updateEvent.GetStatus()
	taskId := status.GetTaskId()
	task, err := s.taskmanager.GetById(taskId)
	if err != nil {
		// The event is from a task that has been deleted from the task manager,
		// ignore updates.
		// NOTE (tim): Do we want to keep deleted task history for a certain amount of time
		// before it's deleted? We would record status updates after it's killed here.
		// ACK update, return.
		status := updateEvent.GetStatus()
		s.scheduler.Acknowledge(status.GetAgentId(), status.GetTaskId(), status.GetUuid())
		return
	}

	state := status.GetState()
	message := status.GetMessage()
	agentId := status.GetAgentId()
	s.taskmanager.Set(state, task)

	switch state {
	case mesos_v1.TaskState_TASK_FAILED:
		s.logger.Emit(logging.ERROR, message)

		// If there's an error, fallback to the regular policy.
		policy, err := s.taskmanager.CheckPolicy(task)
		retryFunc := func() {
			s.taskmanager.Set(manager.UNKNOWN, task)
			s.Scheduler().Revive()
		}
		if err != nil {
			s.logger.Emit(logging.INFO, err.Error())
		}
		s.taskmanager.RunPolicy(policy, retryFunc)
	case mesos_v1.TaskState_TASK_STAGING:
		// NOP, keep task set to "launched".
		s.logger.Emit(logging.INFO, message)
	case mesos_v1.TaskState_TASK_DROPPED:
		// Transient error, we should retry launching. Taskinfo is fine.
		s.logger.Emit(logging.INFO, message)
	case mesos_v1.TaskState_TASK_ERROR:
		// TODO (tim): Error with the taskinfo sent to the agent. Give verbose reasoning back.
		s.logger.Emit(logging.ERROR, message)
	case mesos_v1.TaskState_TASK_FINISHED:
		s.taskmanager.Delete(task)
	case mesos_v1.TaskState_TASK_GONE:
		// Agent is dead and task is lost.
		s.logger.Emit(logging.ERROR, message)
	case mesos_v1.TaskState_TASK_GONE_BY_OPERATOR:
		// Agent might be dead, master is unsure. Will return to RUNNING state possibly or die.
		s.logger.Emit(logging.ERROR, message)
	case mesos_v1.TaskState_TASK_KILLED:
		// Task was killed.
		s.logger.Emit(
			logging.INFO,
			"Task %s on agent %s was killed",
			taskId.GetValue(),
			agentId.GetValue(),
		)
		s.taskmanager.Delete(task)
	case mesos_v1.TaskState_TASK_KILLING:
		// Task is in the process of catching a SIGNAL and shutting down.
		s.logger.Emit(logging.INFO, message)
	case mesos_v1.TaskState_TASK_LOST:
		// Task is unknown to the master and lost. Should reschedule.
		s.logger.Emit(logging.ALARM, "Task %s was lost", taskId.GetValue())
	case mesos_v1.TaskState_TASK_RUNNING:
		s.logger.Emit(logging.INFO, message)
	case mesos_v1.TaskState_TASK_STARTING:
	// Task is still starting up. NOOP
	case mesos_v1.TaskState_TASK_UNKNOWN:
	// Task is unknown to the master. Should ignore.
	case mesos_v1.TaskState_TASK_UNREACHABLE:
		// Agent lost contact with master, could be a network error. No guarantee the task is still running.
		// Should we reschedule after waiting a certain period of time?
		s.logger.Emit(logging.INFO, message)
	default:
		// Somewhere in here the universe started.
	}

	s.scheduler.Acknowledge(agentId, taskId, status.GetUuid())
}

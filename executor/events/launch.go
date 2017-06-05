package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
)

func (d *SprintExecutorController) Launch(launch *exec.Event_Launch) {
	// Send starting state.
	d.executor.Update(&mesos_v1.TaskStatus{TaskId: launch.GetTask().GetTaskId(), State: mesos_v1.TaskState_TASK_STARTING.Enum()})
	// Validate task info.
	//
	// Send update that task is running.
}

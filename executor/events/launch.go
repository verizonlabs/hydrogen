package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/utils"
)

func (d *SprintExecutorController) Launch(launch *exec.Event_Launch) {
	d.logger.Emit(logging.INFO, "Launch event received: ", *launch.Task)
	err := d.executor.Update(&mesos_v1.TaskStatus{
		TaskId: launch.Task.TaskId,
		State:  mesos_v1.TaskState_TASK_RUNNING.Enum(),
		Uuid:   utils.Uuid(),
		Source: mesos_v1.TaskStatus_SOURCE_EXECUTOR.Enum(),
	})
	if err != nil {
		d.logger.Emit(logging.ERROR, "ERROR DURING UPDATE OF TASK", err.Error())
	}
}

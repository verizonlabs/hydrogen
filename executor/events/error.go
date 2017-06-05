package events

import (
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
)

func (d *SprintExecutorController) Error(error *exec.Event_Error) {
	d.logger.Emit(logging.ERROR, error.GetMessage())
}

package events

import (
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
)

func (d *SprintExecutorController) Kill(kill *exec.Event_Kill) {
	d.logger.Emit(logging.INFO, "Kill event received")
}

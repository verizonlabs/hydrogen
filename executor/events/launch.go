package events

import (
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
)

func (d *SprintExecutorController) Launch(launch *exec.Event_Launch) {
	d.logger.Emit(logging.INFO, "Launch event received")
}

package events

import (
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
)

func (d *SprintExecutorController) Subscribed(sub *exec.Event_Subscribed) {
	d.logger.Emit(logging.INFO, "Executor successfully subscribed")
}

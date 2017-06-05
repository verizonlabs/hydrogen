package events

import (
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
)

func (d *SprintExecutorController) Message(message *exec.Event_Message) {
	d.logger.Emit(logging.INFO, "%s", message.GetData())
}

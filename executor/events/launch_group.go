package events

import (
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
)

func (d *SprintExecutorController) LaunchGroup(launchGroup *exec.Event_LaunchGroup) {
	d.logger.Emit(logging.INFO, "Launch_group event received")
}

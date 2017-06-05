package events

import (
	exec "mesos-framework-sdk/include/mesos_v1_executor"
)

func (d *SprintExecutorController) Subscribed(sub *exec.Event_Subscribed) {
	sub.GetFrameworkInfo()
}

package events

import (
	"fmt"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
)

func (d *SprintExecutorController) LaunchGroup(launchGroup *exec.Event_LaunchGroup) {
	fmt.Println(launchGroup.GetTaskGroup())
}

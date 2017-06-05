package events

import (
	"fmt"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
)

func (d *SprintExecutorController) Acknowledged(acknowledge *exec.Event_Acknowledged) {
	// The executor is expected to maintain a list of status updates not acknowledged by the agent via the ACKNOWLEDGE events.
	// The executor is expected to maintain a list of tasks that have not been acknowledged by the agent.
	// A task is considered acknowledged if at least one of the status updates for this task is acknowledged by the agent.
	fmt.Printf("%v\n", acknowledge.GetTaskId())
}

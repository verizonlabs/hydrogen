package events

import (
	"fmt"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
)

func (d *SprintExecutorController) Error(error *exec.Event_Error) {
	fmt.Printf("%v\n", error.GetMessage())
}

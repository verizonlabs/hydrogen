package events

import (
	"fmt"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
)

func (d *SprintExecutorController) Kill(kill *exec.Event_Kill) {
	fmt.Printf("%v, %v\n", kill.GetTaskId(), kill.GetKillPolicy())
}

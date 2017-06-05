package events

import (
	"fmt"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
)

func (d *SprintExecutorController) Message(message *exec.Event_Message) {
	fmt.Printf("%v\n", message.GetData())
}

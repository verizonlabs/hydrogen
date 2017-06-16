package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Message(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
		Data:       []byte(`some message`),
	})
}

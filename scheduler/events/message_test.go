package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

// Test that we can pass a message.
func TestSprintEventController_Message(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
		Data:       []byte(`some message`),
	})
}

// Test if we send an empty message
func TestSprintEventController_MessageNoData(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
	})
}

// Test what we do if we get a nil message
func TestSprintEventController_NilMessage(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Message(nil)
}

// Test if we get a nil agent or nil value within the agent protobuf.
func TestSprintEventController_MessageWithNoAgent(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: nil},
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
		Data:       []byte(`some message`),
	})
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    nil,
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
		Data:       []byte(`some message`),
	})
}

// Test if we get a nil executor or nil value inside the protobuf.
func TestSprintEventController_MessageWithNoExecutor(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: nil},
		Data:       []byte(`some message`),
	})
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: nil,
		Data:       []byte(`some message`),
	})
}

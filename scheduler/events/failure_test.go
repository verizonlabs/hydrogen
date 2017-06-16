package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Failure(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Failure(&mesos_v1_scheduler.Event_Failure{
		AgentId: &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
	})
}

func TestSprintEventController_FailureWithNoAgentID(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Failure(&mesos_v1_scheduler.Event_Failure{
		AgentId: &mesos_v1.AgentID{Value: nil},
	})
	ctrl.Failure(nil)
}

// Copyright 2017 Verizon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

// Test that we can pass a message.
func TestSprintEventController_Message(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
		Data:       []byte(`some message`),
	})
}

// Test if we send an empty message
func TestSprintEventController_MessageNoData(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
	})
}

// Test what we do if we get a nil message
func TestSprintEventController_NilMessage(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Message(nil)
}

// Test if we get a nil agent or nil value within the agent protobuf.
func TestSprintEventController_MessageWithNoAgent(t *testing.T) {
	ctrl := workingEventController()
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
	ctrl := workingEventController()
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

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
	mockLogger "mesos-framework-sdk/logging/test"
	mockResourceManager "mesos-framework-sdk/resources/manager/test"
	sched "mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"hydrogen/scheduler"
	mockTaskManager "hydrogen/task/manager/test"
	mockStorage "hydrogen/task/persistence/test"
	"testing"
)

func TestHandler_Subscribe(t *testing.T) {
	e := NewHandler(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)
	e.Subscribed(&mesos_v1_scheduler.Event_Subscribed{FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("id")}})
}

func TestHandler_Subscribed(t *testing.T) {
	ch := make(chan *mesos_v1_scheduler.Event)

	ch <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}
}

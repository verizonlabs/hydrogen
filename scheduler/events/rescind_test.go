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
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	mockLogger "github.com/verizonlabs/mesos-framework-sdk/logging/test"
	mockResourceManager "github.com/verizonlabs/mesos-framework-sdk/resources/manager/test"
	sched "github.com/verizonlabs/mesos-framework-sdk/scheduler/test"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"github.com/verizonlabs/hydrogen/scheduler"
	mockTaskManager "github.com/verizonlabs/hydrogen/task/manager/test"
	mockStorage "github.com/verizonlabs/hydrogen/task/persistence/test"
	"testing"
)

func TestHandler_Rescind(t *testing.T) {
	e := NewHandler(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)
	e.Rescind(&mesos_v1_scheduler.Event_Rescind{})
}

func TestHandler_RescindWithNil(t *testing.T) {
	e := NewHandler(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)
	e.Rescind(&mesos_v1_scheduler.Event_Rescind{OfferId: nil})
	e.Rescind(nil)
}

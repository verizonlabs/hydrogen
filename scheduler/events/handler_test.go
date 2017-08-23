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
	"mesos-framework-sdk/logging"
	mockLogger "mesos-framework-sdk/logging/test"
	mockResourceManager "mesos-framework-sdk/resources/manager/test"
	sdkScheduler "mesos-framework-sdk/scheduler"
	sched "mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/task/manager"
	"sprint/scheduler"
	"sprint/scheduler/controller"
	"sprint/scheduler/ha"
	mockTaskManager "sprint/task/manager/test"
	"sprint/task/persistence"
	mockStorage "sprint/task/persistence/test"
	"testing"
	"time"
)

// Creates as new working event controller.
func workingEventController() *controller.EventController {
	var (
		cfg = &scheduler.Configuration{
			Leader: &scheduler.LeaderConfiguration{
				IP: "1", // Make sure we break out of our HA loop by matching on what mock storage gives us.
			},
			Executor:  &scheduler.ExecutorConfiguration{},
			Scheduler: &scheduler.SchedulerConfiguration{ReconcileInterval: time.Nanosecond},
			Persistence: &scheduler.PersistenceConfiguration{
				MaxRetries: 0,
			},
		}
		sh sdkScheduler.Scheduler = sched.MockScheduler{}
		m  manager.TaskManager    = &mockTaskManager.MockTaskManager{}
		s  persistence.Storage    = &mockStorage.MockStorage{}
		l  logging.Logger         = &mockLogger.MockLogger{}
		ha                        = ha.NewHA(s, l, cfg.Leader)
	)
	return controller.NewEventController(
		cfg,
		sh,
		m,
		s,
		l,
		ha,
	)
}

// Tests creation of a new event.
func TestHandler_NewHandler(t *testing.T) {
	e := NewHandler(new(mockResourceManager.MockResourceManager))
	if e == nil {
		t.FailNow()
	}
}

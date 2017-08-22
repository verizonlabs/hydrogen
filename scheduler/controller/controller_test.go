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

package controller

import (
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	mockLogger "mesos-framework-sdk/logging/test"
	sdkScheduler "mesos-framework-sdk/scheduler"
	sched "mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/task/manager"
	"os"
	"sprint/scheduler"
	"sprint/scheduler/ha"
	mockTaskManager "sprint/task/manager/test"
	"sprint/task/persistence"
	mockStorage "sprint/task/persistence/test"
	"testing"
	"time"
)

// TODO (tim): Factory function that takes in a list of broken items,
// generates an event controller with those items broken.

// Creates a new working event controller.
func workingEventController() *EventController {
	var (
		cfg *scheduler.Configuration = &scheduler.Configuration{
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
	return NewEventController(
		cfg,
		sh,
		m,
		s,
		l,
		ha,
	)
}

func brokenSchedulerEventController() *EventController {
	var (
		cfg *scheduler.Configuration = &scheduler.Configuration{
			Leader:    &scheduler.LeaderConfiguration{},
			Executor:  &scheduler.ExecutorConfiguration{},
			Scheduler: &scheduler.SchedulerConfiguration{ReconcileInterval: time.Nanosecond},
			Persistence: &scheduler.PersistenceConfiguration{
				MaxRetries: 0,
			},
		}
		sh sdkScheduler.Scheduler = sched.MockBrokenScheduler{}
		m  manager.TaskManager    = &mockTaskManager.MockTaskManager{}
		s  persistence.Storage    = &mockStorage.MockStorage{}
		l  logging.Logger         = &mockLogger.MockLogger{}
		ha                        = ha.NewHA(s, l, cfg.Leader)
	)
	return NewEventController(
		cfg,
		sh,
		m,
		s,
		l,
		ha,
	)
}

// Tests creation of a new event controller.
func TestNewEventController(t *testing.T) {
	ctrl := workingEventController()

	if ctrl == nil {
		t.FailNow()
	}
}

// Verify that we can successfully run our scheduler.
func TestEventController_Run(t *testing.T) {
	ctrl := workingEventController()

	go ctrl.Run(make(chan *mesos_v1_scheduler.Event))
}

// Ensure our controller registers signal handlers and check that they work.
func TestEventController_SignalHandler(t *testing.T) {
	ctrl := workingEventController()
	ctrl.registerShutdownHandlers()
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(os.Interrupt)
	p.Wait()
}

// Test our periodic reconciling.
func TestEventController_periodicReconcile(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.periodicReconcile()
	broken := brokenSchedulerEventController()
	go broken.periodicReconcile()
}

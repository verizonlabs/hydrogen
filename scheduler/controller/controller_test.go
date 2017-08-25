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
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	mockLogger "mesos-framework-sdk/logging/test"
	mockResourceManager "mesos-framework-sdk/resources/manager/test"
	sdkScheduler "mesos-framework-sdk/scheduler"
	sched "mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/task/manager"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"sprint/scheduler"
	"sprint/scheduler/events"
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
	ch := make(chan *mesos_v1_scheduler.Event)
	r := mockResourceManager.MockResourceManager{}
	v := make(chan *sdkTaskManager.Task)
	h := events.NewHandler(ctrl.taskManager, r, ctrl.config, ctrl.scheduler, ctrl.storage, v, ctrl.logger)
	go ctrl.Run(ch, v, h)
}

// Test our periodic reconciling.
func TestEventController_periodicReconcile(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.periodicReconcile()
	broken := brokenSchedulerEventController()
	go broken.periodicReconcile()
}

func TestEventController_listen(t *testing.T) {
	ch := make(chan *mesos_v1_scheduler.Event)
	ctrl := workingEventController()
	r := mockResourceManager.MockResourceManager{}
	v := make(chan *sdkTaskManager.Task)
	h := events.NewHandler(ctrl.taskManager, r, ctrl.config, ctrl.scheduler, ctrl.storage, v, ctrl.logger)
	go ctrl.Run(ch, v, h)

	ch <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}

	// Test all event messages.
	ch <- &mesos_v1_scheduler.Event{
		Type:  mesos_v1_scheduler.Event_ERROR.Enum(),
		Error: &mesos_v1_scheduler.Event_Error{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_FAILURE.Enum(),
		Failure: &mesos_v1_scheduler.Event_Failure{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:          mesos_v1_scheduler.Event_INVERSE_OFFERS.Enum(),
		InverseOffers: &mesos_v1_scheduler.Event_InverseOffers{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_MESSAGE.Enum(),
		Message: &mesos_v1_scheduler.Event_Message{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:   mesos_v1_scheduler.Event_OFFERS.Enum(),
		Offers: &mesos_v1_scheduler.Event_Offers{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: &mesos_v1_scheduler.Event_Rescind{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:   mesos_v1_scheduler.Event_UPDATE.Enum(),
		Update: &mesos_v1_scheduler.Event_Update{},
	}
}

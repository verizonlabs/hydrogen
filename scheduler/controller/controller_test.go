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
	"mesos-framework-sdk/logging"
	mockLogger "mesos-framework-sdk/logging/test"
	sdkScheduler "mesos-framework-sdk/scheduler"
	sched "mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/task/manager"
	"os"
	"sprint/scheduler"
	mockTaskManager "sprint/task/manager/test"
	"sprint/task/persistence"
	mockStorage "sprint/task/persistence/test"
	"testing"
	"time"
)

var (
	states []*mesos_v1.TaskState = []*mesos_v1.TaskState{
		mesos_v1.TaskState_TASK_RUNNING.Enum(),
		mesos_v1.TaskState_TASK_STARTING.Enum(),
		mesos_v1.TaskState_TASK_FINISHED.Enum(),
		mesos_v1.TaskState_TASK_DROPPED.Enum(),
		mesos_v1.TaskState_TASK_ERROR.Enum(),
		mesos_v1.TaskState_TASK_FAILED.Enum(),
		mesos_v1.TaskState_TASK_GONE.Enum(),
		mesos_v1.TaskState_TASK_GONE_BY_OPERATOR.Enum(),
		mesos_v1.TaskState_TASK_UNREACHABLE.Enum(),
		mesos_v1.TaskState_TASK_UNKNOWN.Enum(),
		mesos_v1.TaskState_TASK_STAGING.Enum(),
		mesos_v1.TaskState_TASK_KILLED.Enum(),
		mesos_v1.TaskState_TASK_KILLING.Enum(),
		mesos_v1.TaskState_TASK_ERROR.Enum(),
		mesos_v1.TaskState_TASK_LOST.Enum(),
	}
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
	)
	return &EventController{
		Config:      cfg,
		Scheduler:   sh,
		TaskManager: m,
		Storage:     s,
		Logger:      l,
	}
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
	)
	return &EventController{
		Config:      cfg,
		Scheduler:   sh,
		TaskManager: m,
		Storage:     s,
		Logger:      l,
	}
}

func TestNewSprintEventController(t *testing.T) {
	ctrl := workingEventController()
	if ctrl == nil {
		t.FailNow()
	}
}

/*
func TestEventController_Run(t *testing.T) {
	ctrl := workingEventController()

	go ctrl.Run()

	// Subscribe.
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}
}

// This should utilize the broken factory
func TestEventController_FailureToRun(t *testing.T) {
	ctrl := workingEventController()

	go ctrl.Run()

	// Subscribe.
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}
}*/

func TestEventController_SignalHandler(t *testing.T) {
	ctrl := workingEventController()
	ctrl.registerShutdownHandlers()
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(os.Interrupt)
	p.Wait()
}

func TestEventController_periodicReconcile(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.periodicReconcile()
	broken := brokenSchedulerEventController()
	go broken.periodicReconcile()

}

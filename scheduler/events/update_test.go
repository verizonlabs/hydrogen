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

var states []*mesos_v1.TaskState = []*mesos_v1.TaskState{
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

func TestHandler_Update(t *testing.T) {
	e := NewHandler(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)

	for _, state := range states {
		e.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
				State:  state,
			}})
	}
}

func TestHandler_UpdateWithNilTaskId(t *testing.T) {
	e := NewHandler(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)

	for _, state := range states {
		e.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: nil},
				State:  state,
			}})

		e.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: nil,
				State:  state,
			}})
	}

	e.Update(nil)
}

func TestHandler_UpdateWithInvalidState(t *testing.T) {
	e := NewHandler(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)

	e.Update(&mesos_v1_scheduler.Event_Update{
		Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: nil},
			State:  nil,
		}})

	e.Update(&mesos_v1_scheduler.Event_Update{
		Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: nil},
			State:  mesos_v1.TaskState(100).Enum(), // Doesn't exist.
		}})
}

func TestHandler_UpdateWith(t *testing.T) {
	e := NewHandler(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)

	for _, state := range states {
		e.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
				State:  state,
			}})
	}
}

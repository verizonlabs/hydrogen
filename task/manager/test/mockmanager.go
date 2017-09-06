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

package test

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/structures/test"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/task/retry"
	"time"
)

type MockTaskManager struct{}

func (m MockTaskManager) Add(...*manager.Task) error {
	return nil
}

func (m MockTaskManager) Restore(*manager.Task) {}

func (m MockTaskManager) Delete(...*manager.Task) error {
	return nil
}

func (m MockTaskManager) GetGroup(*manager.Task) ([]*manager.Task, error) {
	return nil, nil
}

func (m MockTaskManager) Get(*string) (*manager.Task, error) {
	return &manager.Task{Retry: &retry.TaskRetry{
		TotalRetries: 0,
		MaxRetries:   0,
		RetryTime:    time.Nanosecond,
		Backoff:      true,
		Name:         "id",
	}}, nil
}

func (m MockTaskManager) GetById(id *mesos_v1.TaskID) (*manager.Task, error) {
	return &manager.Task{Retry: &retry.TaskRetry{
		TotalRetries: 0,
		MaxRetries:   0,
		RetryTime:    time.Nanosecond,
		Backoff:      true,
		Name:         "id",
	}}, nil
}

func (m MockTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockTaskManager) Update(...*manager.Task) error {
	return nil
}

func (m MockTaskManager) AllByState(state mesos_v1.TaskState) ([]*manager.Task, error) {
	return []*manager.Task{}, nil
}

func (m MockTaskManager) TotalTasks() int {
	return 0
}

func (m MockTaskManager) All() ([]*manager.Task, error) {
	return []*manager.Task{manager.NewTask(
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
		[]task.Filter{},
		&retry.TaskRetry{},
		1,
		manager.GroupInfo{})}, nil
}

//
// Mock Broken Task Manager
//
type MockBrokenTaskManager struct{}

var (
	broken error = errors.New("Broken")
)

func (m MockBrokenTaskManager) Add(...*manager.Task) error {
	return broken
}

func (m MockBrokenTaskManager) Restore(*manager.Task) {}

func (m MockBrokenTaskManager) Delete(...*manager.Task) error {
	return broken
}

func (m MockBrokenTaskManager) GetGroup(*manager.Task) ([]*manager.Task, error) {
	return nil, nil
}

func (m MockBrokenTaskManager) Get(*string) (*manager.Task, error) {
	return nil, broken
}

func (m MockBrokenTaskManager) GetById(id *mesos_v1.TaskID) (*manager.Task, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockBrokenTaskManager) Update(...*manager.Task) error {
	return errors.New("Broken.")
}

func (m MockBrokenTaskManager) State(name *string) (*mesos_v1.TaskState, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) TotalTasks() int {
	return 0
}

func (m MockBrokenTaskManager) Tasks() structures.DistributedMap {
	return &test.MockBrokenDistributedMap{}
}

func (m MockBrokenTaskManager) All() ([]*manager.Task, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) AllByState(state mesos_v1.TaskState) ([]*manager.Task, error) {
	return nil, errors.New("Broken.")
}

type MockTaskManagerQueued struct{}

func (m MockTaskManagerQueued) Add(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) Restore(*manager.Task) {}

func (m MockTaskManagerQueued) Delete(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) GetGroup(*manager.Task) ([]*manager.Task, error) {
	return []*manager.Task{}, nil
}

func (m MockTaskManagerQueued) Get(*string) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManagerQueued) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManagerQueued) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockTaskManagerQueued) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) State(name *string) (*mesos_v1.TaskState, error) {
	return new(mesos_v1.TaskState), nil
}

func (m MockTaskManagerQueued) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return []*mesos_v1.TaskInfo{{}}, nil
}

func (m MockTaskManagerQueued) TotalTasks() int {
	return 1
}

func (m MockTaskManagerQueued) Tasks() structures.DistributedMap {
	return &test.MockDistributedMap{}
}

func (m MockTaskManagerQueued) All() ([]*manager.Task, error) {
	return []*manager.Task{manager.NewTask(
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
		[]task.Filter{},
		&retry.TaskRetry{},
		1,
		manager.GroupInfo{})}, nil
}

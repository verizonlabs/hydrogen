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
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/task/retry"
)

type (
	MockApiManager       struct{}
	MockBrokenApiManager struct{}
)

func (m MockApiManager) Deploy([]byte) ([]*manager.Task, error) {
	return []*manager.Task{
		{Info: &mesos_v1.TaskInfo{}},
	}, nil
}
func (m MockApiManager) Kill([]byte) (string, error) { return "", nil }
func (m MockApiManager) Update([]byte) ([]*manager.Task, error) {
	return []*manager.Task{{Info: &mesos_v1.TaskInfo{}}}, nil
}

func (m MockApiManager) Status(string) (*manager.Task, error) {
	return &manager.Task{}, nil
}
func (m MockApiManager) AllTasks() ([]*manager.Task, error) {
	return []*manager.Task{manager.NewTask(
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
		[]task.Filter{},
		&retry.TaskRetry{},
		1,
		manager.GroupInfo{})}, nil
}

func (m MockBrokenApiManager) Deploy([]byte) ([]*manager.Task, error) {
	return nil, errors.New("Broken")
}
func (m MockBrokenApiManager) Kill([]byte) (string, error) { return "", errors.New("Broken") }
func (m MockBrokenApiManager) Update([]byte) ([]*manager.Task, error) {
	return nil, errors.New("Broken")
}
func (m MockBrokenApiManager) Status(string) (*manager.Task, error) {
	return &manager.Task{}, errors.New("Broken")
}
func (m MockBrokenApiManager) AllTasks() ([]*manager.Task, error) {
	return nil, errors.New("Broken")
}

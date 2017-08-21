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

package manager

import (
	"encoding/json"
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	r "mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/task"
	t "mesos-framework-sdk/task/manager"
	"sprint/task/builder"
)

type (
	ApiParser interface {
		Deploy([]byte) ([]*t.Task, error)
		Kill([]byte) (string, error)
		Update([]byte) ([]*t.Task, error)
		Status(string) (mesos_v1.TaskState, error)
		AllTasks() ([]*t.Task, error)
	}

	Parser struct {
		//		ctrlPlane       control.ControlPlane
		resourceManager r.ResourceManager
		taskManager     t.TaskManager
		scheduler       scheduler.Scheduler
	}
)

// NewApiParser returns an object that marshalls JSON and handles the input from the API endpoints.
func NewApiParser(r r.ResourceManager, t t.TaskManager, s scheduler.Scheduler) *Parser {
	return &Parser{
		resourceManager: r,
		taskManager:     t,
		scheduler:       s,
	}
}

// Deploy takes a slice of bytes and marshals them into a Application json struct.
func (m *Parser) Deploy(decoded []byte) ([]*t.Task, error) {
	var appJSON []*task.ApplicationJSON
	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return nil, err
	}

	if len(appJSON) == 0 {
		return nil, errors.New("No valid application passed in.")
	}

	mesosTasks, err := builder.Application(appJSON...)
	if err != nil {
		return nil, err
	}

	err = m.taskManager.Add(mesosTasks...)
	if err != nil {
		return nil, err
	}

	m.scheduler.Revive()
	return mesosTasks, nil
}

// Update takes a slice of bytes and marshalls them into an ApplicationJSON struct.
func (m *Parser) Update(decoded []byte) ([]*t.Task, error) {
	var appJSON task.ApplicationJSON
	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return nil, err
	}

	taskToKill, err := m.taskManager.Get(&appJSON.Name)
	if err != nil {
		return nil, err
	}

	mesosTask, err := builder.Application(&appJSON)
	if err != nil {
		return nil, err
	}

	m.scheduler.Kill(taskToKill.Info.GetTaskId(), taskToKill.Info.GetAgentId())
	m.taskManager.Add(mesosTask...)
	m.scheduler.Revive()

	return mesosTask, nil
}

// Kill takes a slice of bytes and marshalls them into a kill json struct.
func (m *Parser) Kill(decoded []byte) (string, error) {
	var appJSON task.KillJson
	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return "", err
	}

	// Make sure we have a name to look up
	if appJSON.Name == nil {
		return "", errors.New("Task name is nil")
	}

	// Look up task in task manager
	tsk, err := m.taskManager.Get(appJSON.Name)
	if err != nil {
		return "", err
	}

	err = m.taskManager.Delete(tsk)
	if err != nil {
		return "", err
	}

	// If we are "unknown" that means the master doesn't know about the task, no need to make an HTTP call.
	if tsk.State != t.UNKNOWN {
		_, err := m.scheduler.Kill(tsk.Info.GetTaskId(), tsk.Info.GetAgentId())
		if err != nil {
			return "", err
		}
	}

	return *appJSON.Name, nil
}

func (m *Parser) Status(name string) (mesos_v1.TaskState, error) {
	tsk, err := m.taskManager.Get(&name)
	if err != nil {
		return t.UNKNOWN, err
	}

	return tsk.State, nil
}

func (m *Parser) AllTasks() ([]*t.Task, error) {
	tasks, err := m.taskManager.All()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

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
	"fmt"
	"github.com/verizonlabs/hydrogen/task/builder"
	r "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/task"
	t "github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/plan"
	"os"
)

type (
	ApiParser interface {
		Deploy([]byte) ([]*t.Task, error)
		Kill([]byte) (string, error)
		Update([]byte) ([]*t.Task, error)
		Status(string) (*t.Task, error)
		AllTasks() ([]*t.Task, error)
	}

	factory func(tasks []*manager.Task, planType PlanType) Plan

	Parser struct {
		planFactory factory
		planner     plan.PlanQueue
	}
)

// NewApiParser returns an object that marshalls JSON and handles the input from the API endpoints.
func NewApiParser(planner plan.PlanQueue, planFactory factory) *Parser {
	return &Parser{
		planner:     planner,
		planFactory: planFactory,
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

	// Create a Launch plan.
	m.planner.Push(m.planFactory(mesosTasks, plan.Launch))

	return mesosTasks, nil
}

// Update takes a slice of bytes and marshalls them into an ApplicationJSON struct.
func (m *Parser) Update(decoded []byte) ([]*t.Task, error) {
	var appJSON []*task.ApplicationJSON

	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return nil, err
	}

	if len(appJSON) == 0 {
		return nil, errors.New("No valid application passed in.")
	}

	mesosTask, err = builder.Application(appJSON...)
	if err != nil {
		return nil, err
	}

	m.planner.Push(m.planFactory(mesosTask, plan.Update))

	return mesosTask, nil
}

// Kill takes a slice of bytes and marshalls them into a kill json struct.
func (m *Parser) Kill(decoded []byte) (string, error) {
	var appJSON []*task.KillJson
	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return "", err
	}

	if len(appJSON) == 0 {
		return "", errors.New("No valid tasks passed in.")
	}

	// We only have task names instaed of tasks., create a kill plan with those names.

	return *appJSON[0].Name, nil
}

func (m *Parser) Status(name string) (*t.Task, error) {
	tsk, err := m.taskManager.Get(&name)
	if err != nil {
		return nil, err
	}

	return tsk, nil
}

func (m *Parser) AllTasks() ([]*t.Task, error) {
	tasks, err := m.taskManager.All()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

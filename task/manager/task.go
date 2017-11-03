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
	"errors"
	"github.com/verizonlabs/hydrogen/task/persistence"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
	"github.com/verizonlabs/mesos-framework-sdk/structures"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"github.com/verizonlabs/mesos-framework-sdk/utils"
	"strconv"
	"strings"
	"sync"
)

const (
	// Root directory
	TASK_DIRECTORY = "/tasks/"
)

type (
	// Our primary task handler that implements the above interface.
	// The task handler manages all tasks that are submitted, updated, or deleted.
	// Offers from Mesos are matched up with user-submitted tasks, and those tasks are updated via event callbacks.
	TaskHandler struct {
		mutex   sync.RWMutex
		tasks   map[string]*manager.Task
		groups  map[string][]*mesos_v1.AgentID
		storage persistence.Storage
		retries structures.DistributedMap
		logger  logging.Logger
	}
)

// Returns the core task manager that's used by the scheduler.
func NewTaskManager(
	cmap map[string]*manager.Task,
	storage persistence.Storage,
	logger logging.Logger) manager.TaskManager {

	handler := &TaskHandler{
		tasks:   cmap,
		storage: storage,
		retries: structures.NewConcurrentMap(),
		groups:  make(map[string][]*mesos_v1.AgentID),
		logger:  logger,
	}

	return handler
}

// Add and persists a new task into the task manager.
// Duplicate task names are not allowed by Mesos, thus they are not allowed here.
func (m *TaskHandler) Add(tasks ...*manager.Task) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, t := range tasks {
		t.State = manager.UNKNOWN

		originalName := t.Info.GetName()
		taskId := t.Info.GetTaskId().GetValue()

		// If we have a single instance, only add it.
		if t.Instances == 1 {
			if _, ok := m.tasks[t.Info.GetName()]; ok {
				return errors.New("Task " + t.Info.GetName() + " already exists")
			}
			// Write forward.
			data, err := t.Encode()
			if err != nil {
				return err
			}

			err = m.storageWrite(t, data)
			if err != nil {
				m.logger.Emit(logging.ERROR, "Storage error: %v", err)
				return err
			}
			m.tasks[t.Info.GetName()] = t
			continue
		}

		// Add a group
		t.GroupInfo = manager.GroupInfo{GroupName: t.Info.GetName() + "/", InGroup: true}

		for i := 0; i < t.Instances; i++ {
			// TODO(tim): t.Copy(Name, TaskId)
			duplicate := *t
			tmp := *t.Info // Make a copy
			duplicate.Info = &tmp
			duplicate.Info.Name = utils.ProtoString(originalName + "-" + strconv.Itoa(i+1))
			duplicate.Info.TaskId = &mesos_v1.TaskID{Value: utils.ProtoString(taskId + "-" + strconv.Itoa(i+1))}
			if _, ok := m.tasks[duplicate.Info.GetName()]; ok {
				return errors.New("Task " + duplicate.Info.GetName() + " already exists")
			}

			// Write forward.
			data, err := duplicate.Encode()
			if err != nil {
				return err
			}

			err = m.storageWrite(&duplicate, data)
			if err != nil {
				m.logger.Emit(logging.ERROR, "Storage error: %v", err)
				return err
			}
			m.tasks[duplicate.Info.GetName()] = &duplicate
		}
	}

	return nil
}

func (m *TaskHandler) Restore(task *manager.Task) {
	m.tasks[task.Info.GetName()] = task
}

// Delete a task from memory and etcd, and clears any associated policy.
func (m *TaskHandler) Delete(tasks ...*manager.Task) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, t := range tasks {
		// Try to delete it from storage first.
		if err := m.storageDelete(t); err != nil {
			return err
		}
		// Then from in memory.
		delete(m.tasks, t.Info.GetName())

	}
	return nil
}

// Get a task by its name.
func (m *TaskHandler) Get(name *string) (*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if response, ok := m.tasks[*name]; ok {
		return response, nil
	}
	return nil, errors.New(*name + " not found.")
}

func (m *TaskHandler) GetGroup(task *manager.Task) ([]*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if !task.GroupInfo.InGroup {
		return nil, errors.New("Task " + task.Info.GetName() + " is not in a group.")
	}
	tasks := make([]*manager.Task, 0, task.Instances)
	for i := 0; i < task.Instances; i++ {
		nameSplit := strings.Split(task.Info.GetName(), "-")
		if t, ok := m.tasks[nameSplit[0]+"-"+strconv.Itoa(i+1)]; ok {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

// GetById : get a task by it's ID.
func (m *TaskHandler) GetById(id *mesos_v1.TaskID) (*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.tasks) == 0 {
		return nil, errors.New("There are no tasks")
	}

	for _, v := range m.tasks {
		if id.GetValue() == v.Info.GetTaskId().GetValue() {
			return v, nil
		}
	}
	return nil, errors.New("Could not find task by id: " + id.GetValue())
}

// HasTask indicates whether or not the task manager holds the specified task.
func (m *TaskHandler) HasTask(task *mesos_v1.TaskInfo) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.tasks) == 0 {
		return false
	}
	if _, ok := m.tasks[task.GetName()]; !ok {
		return false
	}
	return true
}

// TotalTasks the total number of tasks that the task manager holds.
func (m *TaskHandler) TotalTasks() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.tasks)
}

// Update the given task with the given state.
func (m *TaskHandler) Update(tasks ...*manager.Task) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, task := range tasks {
		data, err := task.Encode()
		if err != nil {
			return err
		}
		if err := m.storageWrite(task, data); err != nil {
			return err
		}

		m.tasks[task.Info.GetName()] = task
	}

	return nil
}

// AllByState gets all tasks that match the given state.
func (m *TaskHandler) AllByState(state mesos_v1.TaskState) ([]*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}

	tasks := []*manager.Task{}
	for _, v := range m.tasks {
		if v.State == state {
			tasks = append(tasks, v)
		}
	}

	if len(tasks) == 0 {
		return nil, errors.New("No tasks with state " + state.String())
	}

	return tasks, nil
}

// All gets all tasks that are known to the task manager.
func (m *TaskHandler) All() ([]*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}
	allTasks := []*manager.Task{}
	for _, t := range m.tasks {
		allTasks = append(allTasks, t)
	}
	return allTasks, nil
}

// Function that wraps writing to the storage backend.
func (m *TaskHandler) storageWrite(task *manager.Task, encoded []byte) error {
	var writeKey string
	var id string = task.Info.GetTaskId().GetValue()
	var name string = task.Info.GetName()
	if task.GroupInfo.InGroup {
		writeKey = TASK_DIRECTORY + task.GroupInfo.GroupName + id
	} else {
		writeKey = TASK_DIRECTORY + id
	}
	err := m.storage.Update(writeKey, string(encoded))
	if err != nil {
		m.logger.Emit(
			logging.ERROR, "Failed to update task %s with name %s to persistent data store. Retrying...",
			id,
			name,
		)
	}
	return err
}

// Function that wraps deleting from the storage backend.
func (m *TaskHandler) storageDelete(task *manager.Task) error {
	var del string
	var id string = task.Info.GetTaskId().GetValue()
	if task.GroupInfo.InGroup {
		del = TASK_DIRECTORY + task.GroupInfo.GroupName + id
	} else {
		del = TASK_DIRECTORY + id
	}
	err := m.storage.Delete(del)
	if err != nil {
		m.logger.Emit(logging.ERROR, err.Error())
	}
	return err
}

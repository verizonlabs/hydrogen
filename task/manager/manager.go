package manager

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task/manager"
)

/*
Satisfies the TaskManager interface in sdk

Sprint Task Manager integrates logging and storage backend-
All tasks are written only during creation, updates and deletes.
Reads are reserved for reconciliation calls.
*/

const (
	TASK_DIRECTORY = "/tasks/"
)

type SprintTaskManager struct {
	tasks   *structures.ConcurrentMap
	storage persistence.Storage
	logger  logging.Logger
}

func NewTaskManager(cmap *structures.ConcurrentMap, storage persistence.Storage, logger logging.Logger) manager.TaskManager {
	return &SprintTaskManager{
		tasks:   cmap,
		storage: storage,
		logger:  logger,
	}
}

func (m *SprintTaskManager) encode(task *mesos_v1.TaskInfo, state mesos_v1.TaskState) (bytes.Buffer, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	err := e.Encode(manager.Task{
		Info:  task,
		State: state,
	})

	if err != nil {
		m.logger.Emit(logging.ERROR, "Failed to encode task")
		return b, err
	}

	return b, nil
}

func (m *SprintTaskManager) Add(t *mesos_v1.TaskInfo) error {
	// Write forward.
	encoded, err := m.encode(t, manager.UNKNOWN)
	if err != nil {
		return err
	}
	id := t.TaskId.GetValue()
	if err := m.storage.Create(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes())); err != nil {
		m.logger.Emit(logging.ERROR, "Failed to save task %s with name %s to persistent data store", id, t.GetName())
		return err
	}

	name := t.GetName()
	if m.tasks.Get(name) != nil {
		return errors.New("Task " + name + " already exists")
	}

	m.tasks.Set(t.GetName(), manager.Task{
		State: manager.UNKNOWN,
		Info:  t,
	})

	return nil
}

func (m *SprintTaskManager) Delete(task *mesos_v1.TaskInfo) {
	m.logger.Emit(logging.INFO, TASK_DIRECTORY+task.GetTaskId().GetValue())
	err := m.storage.Delete(TASK_DIRECTORY + task.GetTaskId().GetValue())
	if err != nil {
		m.logger.Emit(logging.ERROR, err.Error())
	}
	m.tasks.Delete(task.GetName())
}

// TODO make param a TaskInfo to be consistent with all other methods that take a TaskInfo.
func (m *SprintTaskManager) Get(name *string) (*mesos_v1.TaskInfo, error) {
	ret := m.tasks.Get(*name)
	if ret != nil {
		return ret.(manager.Task).Info, nil
	}

	return nil, errors.New("Could not find task.")
}

// Check to see if any tasks we have match the id passed in.
func (m *SprintTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	if m.tasks.Length() == 0 {
		return nil, errors.New("Task manager is empty.")
	}

	for v := range m.tasks.Iterate() {
		task := v.Value.(manager.Task)
		if task.Info.GetTaskId().GetValue() == id.GetValue() {
			return task.Info, nil
		}
	}

	return nil, errors.New("Could not find task by id: " + id.GetValue())
}

func (m *SprintTaskManager) HasTask(task *mesos_v1.TaskInfo) bool {
	ret := m.tasks.Get(task.GetName())
	if ret == nil {
		return false
	}

	return true
}

func (m *SprintTaskManager) TotalTasks() int {
	return m.tasks.Length()
}

func (m *SprintTaskManager) Tasks() *structures.ConcurrentMap {
	return m.tasks
}

// Update a task with a certain state.
func (m *SprintTaskManager) Set(state mesos_v1.TaskState, t *mesos_v1.TaskInfo) {
	// Write forward.
	encoded, err := m.encode(t, state)
	if err != nil {
		m.logger.Emit(logging.INFO, err.Error())
	}

	id := t.TaskId.GetValue()
	if err := m.storage.Update(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes())); err != nil {
		m.logger.Emit(logging.ERROR, "Failed to update task %s with name %s to persistent data store", id, t.GetName())
	}

	m.tasks.Set(t.GetName(), manager.Task{
		Info:  t,
		State: state,
	})

	switch state {
	case manager.FINISHED:
		m.Delete(t)
	case manager.KILLED:
		m.Delete(t)
	}
}

// Get's all tasks within a certain state.
func (m *SprintTaskManager) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	tasks := []*mesos_v1.TaskInfo{}
	for v := range m.tasks.Iterate() {
		task := v.Value.(manager.Task)
		if task.State == state {
			tasks = append(tasks, task.Info)
		}
	}

	if len(tasks) == 0 {
		return nil, errors.New("No tasks found with state of " + state.String())
	}

	return tasks, nil
}

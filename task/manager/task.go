package manager

import (
	"encoding/json"
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"sprint/scheduler"
	"sprint/task/persistence"
	"strconv"
	"sync"
)

const (
	// Root directory
	TASK_DIRECTORY = "/tasks/"
	// Groups are appended onto this
	// /tasks/groupName/groupTask1
	// /tasks/groupName/groupTask2...etc
)

type (
	// Our primary task handler that implements the above interface.
	// The task handler manages all tasks that are submitted, updated, or deleted.
	// Offers from Mesos are matched up with user-submitted tasks, and those tasks are updated via event callbacks.
	SprintTaskHandler struct {
		mutex   sync.RWMutex
		tasks   map[string]manager.Task
		groups  map[string][]*mesos_v1.AgentID
		storage persistence.Storage
		retries structures.DistributedMap
		config  *scheduler.Configuration
		logger  logging.Logger
	}
)

// Returns the core task manager that's used by the scheduler.
func NewTaskManager(
	cmap map[string]manager.Task,
	storage persistence.Storage,
	config *scheduler.Configuration,
	logger logging.Logger) manager.TaskManager {

	handler := &SprintTaskHandler{
		tasks:   cmap,
		storage: storage,
		retries: structures.NewConcurrentMap(),
		config:  config,
		groups:  make(map[string][]*mesos_v1.AgentID),
		logger:  logger,
	}

	return handler
}

// Add and persists a new task into the task manager.
// Duplicate task names are not allowed by Mesos, thus they are not allowed here.
func (m *SprintTaskHandler) Add(tasks ...*manager.Task) error {
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
			data, err := m.encode(t)
			if err != nil {
				return err
			}
			err = m.storageWrite(t.Info.GetTaskId().GetValue(), data)
			if err != nil {
				m.logger.Emit(logging.ERROR, "Storage error: %v", err)
				return err
			}
			m.tasks[t.Info.GetName()] = *t
		} else {
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
				data, err := m.encode(&duplicate)
				if err != nil {
					return err
				}

				err = m.storageWrite(duplicate.GroupInfo.GroupName+duplicate.Info.GetTaskId().GetValue(), data)
				if err != nil {
					m.logger.Emit(logging.ERROR, "Storage error: %v", err)
					return err
				}
				m.tasks[duplicate.Info.GetName()] = duplicate
			}
		}
	}

	return nil
}

// Delete a task from memory and etcd, and clears any associated policy.
func (m *SprintTaskHandler) Delete(tasks ...*manager.Task) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, t := range tasks {
		delete(m.tasks, t.Info.GetName())
		if t.GroupInfo.InGroup {
			m.storageDelete(t.GroupInfo.GroupName + t.Info.GetTaskId().GetValue())
		} else {
			m.storageDelete(t.Info.GetTaskId().GetValue())
		}
	}
	return nil
}

// Get a task by its name.
func (m *SprintTaskHandler) Get(name *string) (*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if response, ok := m.tasks[*name]; ok {
		return &response, nil
	}
	return nil, errors.New(*name + " not found.")
}

// GetById : get a task by it's ID.
func (m *SprintTaskHandler) GetById(id *mesos_v1.TaskID) (*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.tasks) == 0 {
		return nil, errors.New("There are no tasks")
	}

	for _, v := range m.tasks {
		if id.GetValue() == v.Info.GetTaskId().GetValue() {
			return &v, nil
		}
	}
	return nil, errors.New("Could not find task by id: " + id.GetValue())
}

// HasTask indicates whether or not the task manager holds the specified task.
func (m *SprintTaskHandler) HasTask(task *mesos_v1.TaskInfo) bool {
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
func (m *SprintTaskHandler) TotalTasks() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.tasks)
}

// Update the given task with the given state.
func (m *SprintTaskHandler) Update(tasks ...*manager.Task) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, task := range tasks {
		data, err := m.encode(task)
		if err != nil {
			return err
		}
		if task.GroupInfo.InGroup {
			if err := m.storageWrite(task.GroupInfo.GroupName+task.Info.GetTaskId().GetValue(), data); err != nil {
				return err
			}
		} else {
			if err := m.storageWrite(task.Info.GetTaskId().GetValue(), data); err != nil {
				return err
			}
		}
		m.tasks[task.Info.GetName()] = *task
	}

	return nil
}

// AllByState gets all tasks that match the given state.
func (m *SprintTaskHandler) AllByState(state mesos_v1.TaskState) ([]*manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}

	tasks := []*manager.Task{}
	for _, v := range m.tasks {
		if v.State == state {
			tasks = append(tasks, &v)
		}
	}

	if len(tasks) == 0 {
		return nil, errors.New("No tasks with state " + state.String())
	}

	return tasks, nil
}

// All gets all tasks that are known to the task manager.
func (m *SprintTaskHandler) All() ([]manager.Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}
	allTasks := []manager.Task{}
	for _, t := range m.tasks {
		allTasks = append(allTasks, t)
	}
	return allTasks, nil
}

// Function that wraps writing to the storage backend.
func (m *SprintTaskHandler) storageWrite(id string, encoded []byte) error {
	err := m.storage.Update(TASK_DIRECTORY+id, string(encoded))
	if err != nil {
		m.logger.Emit(
			logging.ERROR, "Failed to update task %s with name %s to persistent data store. Retrying...",
			id,
		)
	}
	return err
}

// Function that wraps deleting from the storage backend.
func (m *SprintTaskHandler) storageDelete(taskId string) error {
	err := m.storage.Delete(TASK_DIRECTORY + taskId)
	if err != nil {
		m.logger.Emit(logging.ERROR, err.Error())
	}
	return err
}

// Encodes task data.
func (m *SprintTaskHandler) encode(task *manager.Task) ([]byte, error) {
	data, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	return data, err
}

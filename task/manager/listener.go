package manager

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"sprint/scheduler"
	"sprint/task/persistence"
	"sprint/task/retry"
	"sync"
	"time"
)

const (
	TASK_DIRECTORY = "tasks"
)

type (
	// Provides pluggable task managers that the framework can work with.
	// Also used extensively for testing with mocks.
	SprintTaskManager interface {
		manager.TaskManager
		retry.Retry
	}

	// Our primary task handler that implements the above interface.
	// The task handler manages all tasks that are submitted, updated, or deleted.
	// Offers from Mesos are matched up with user-submitted tasks, and those tasks are updated via event callbacks.
	SprintTaskHandler struct {
		mutex   sync.RWMutex
		buffer  *bytes.Buffer
		encoder *gob.Encoder
		tasks   map[string]manager.Task
		groups  map[string][]*mesos_v1.AgentID
		storage persistence.Storage
		retries structures.DistributedMap
		config  *scheduler.Configuration
		logger  logging.Logger
	}
)

var (
	DEFAULT_RETRY_POLICY = &task.TimeRetry{
		Time:       "1.5",
		Backoff:    true,
		MaxRetries: 3,
	}
)

// Returns the core task manager that's used by the scheduler.
func NewTaskManager(
	cmap map[string]manager.Task,
	storage persistence.Storage,
	config *scheduler.Configuration,
	logger logging.Logger) SprintTaskManager {

	b := new(bytes.Buffer)
	handler := &SprintTaskHandler{
		buffer:  b,
		encoder: gob.NewEncoder(b),
		tasks:   cmap,
		storage: storage,
		retries: structures.NewConcurrentMap(),
		config:  config,
		groups:  make(map[string][]*mesos_v1.AgentID),
		logger:  logger,
	}
	return handler
}

//
// Methods to satisfy retry interface.
//
func (s *SprintTaskHandler) AddPolicy(policy *task.TimeRetry, mesosTask *mesos_v1.TaskInfo) error {
	if mesosTask != nil {
		if policy == nil {
			policy = &task.TimeRetry{
				Time:       "1.5",
				Backoff:    true,
				MaxRetries: 3,
			}
		}

		t, err := time.ParseDuration(policy.Time + "s")
		if err != nil {
			return errors.New("Invalid time given in policy.")
		}

		s.retries.Set(mesosTask.GetName(), &retry.TaskRetry{
			TotalRetries: 0,
			RetryTime:    t,
			Backoff:      policy.Backoff,
			Name:         mesosTask.GetName(),
		})
		return nil
	}
	return errors.New("Nil mesos task passed in")
}

// Runs the specified policy with a configurable backoff.
func (s *SprintTaskHandler) RunPolicy(policy *retry.TaskRetry, f func() error) error {
	policy.TotalRetries += 1 // Increment retry counter.

	// Minimum is 1 seconds, max is 60.
	if policy.RetryTime < 1*time.Second {
		policy.RetryTime = 1 * time.Second
	} else if policy.RetryTime > time.Minute {
		policy.RetryTime = time.Minute
	}

	delay := policy.RetryTime + policy.RetryTime

	// Total backoff can't be greater than 5 minutes.
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}

	policy.RetryTime = delay // update with new time.
	go func() {
		time.Sleep(delay)
		f()
	}()

	return nil
}

// Checks whether a policy exists for a given task.
func (s *SprintTaskHandler) CheckPolicy(mesosTask *mesos_v1.TaskInfo) *retry.TaskRetry {
	if mesosTask != nil {
		policy := s.retries.Get(mesosTask.GetName())
		if policy != nil {
			return policy.(*retry.TaskRetry)
		}
	}

	return nil
}

// Removes an existing policy associated with the given task.
func (s *SprintTaskHandler) ClearPolicy(mesosTask *mesos_v1.TaskInfo) error {
	if mesosTask != nil {
		s.retries.Delete(mesosTask.GetTaskId().GetValue())
		return nil
	}
	return errors.New("Nil mesos task passed in")
}

// Add and persists a new task into the task manager.
// Duplicate task names are not allowed by Mesos, thus they are not allowed here.
func (m *SprintTaskHandler) Add(task *mesos_v1.TaskInfo) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Use a unique ID for storing in the map, taskid?
	if _, ok := m.tasks[task.GetName()]; ok {
		return errors.New("Task " + task.GetName() + " already exists")
	}

	// Write forward.
	if err := m.encode(task, manager.UNKNOWN); err != nil {
		return err
	}

	id := task.TaskId.GetValue()

	policy := m.storage.CheckPolicy(nil)

	if err := m.storage.RunPolicy(policy, m.storageWrite(id, m.buffer)); err != nil {
		return err
	}

	m.tasks[task.GetName()] = manager.Task{
		State: manager.UNKNOWN,
		Info:  task,
	}

	return nil
}

// Delete a task from memory and etcd, and clears any associated policy.
func (m *SprintTaskHandler) Delete(task *mesos_v1.TaskInfo) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	policy := m.storage.CheckPolicy(nil)

	if err := m.storage.RunPolicy(policy, m.storageDelete(task.GetTaskId().GetValue())); err != nil {
		return err
	}

	delete(m.tasks, task.GetName())
	m.ClearPolicy(task)

	return nil
}

// Get a task by its name.
func (m *SprintTaskHandler) Get(name *string) (*mesos_v1.TaskInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if response, ok := m.tasks[*name]; ok {
		return response.Info, nil
	}
	return nil, errors.New("Could not find task.")
}

// GetById : get a task by it's ID.
func (m *SprintTaskHandler) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.tasks) == 0 {
		// TODO (tim): Standardize Error messages.
		return nil, errors.New("There are no tasks")
	}
	for _, v := range m.tasks {
		if id.GetValue() == v.Info.GetTaskId().GetValue() {
			return v.Info, nil
		}
	}
	return nil, errors.New("Could not find task by id: " + id.GetValue())
}

// HasTask indicates whether or not the task manager holds the specified task.
func (m *SprintTaskHandler) HasTask(task *mesos_v1.TaskInfo) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.tasks) == 0 {
		return false // If the manager is empty, this one doesn't exist.
	}
	if _, ok := m.tasks[task.GetName()]; !ok {
		return false
	}
	return true
}

// TotalTasks the total number of tasks that the task manager holds.
func (m *SprintTaskHandler) TotalTasks() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.tasks)
}

// Update the given task with the given state.
func (m *SprintTaskHandler) Set(state mesos_v1.TaskState, t *mesos_v1.TaskInfo) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if err := m.encode(t, state); err != nil {
		return err
	}

	id := t.GetTaskId().GetValue()

	policy := m.storage.CheckPolicy(nil)

	if err := m.storage.RunPolicy(policy, m.storageWrite(id, m.buffer)); err != nil {
		return err
	}

	m.tasks[t.GetName()] = manager.Task{
		Info:  t,
		State: state,
	}

	return nil
}

// Gets the task's state based on the supplied task name.
func (m *SprintTaskHandler) State(name *string) (*mesos_v1.TaskState, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if r, ok := m.tasks[*name]; ok {
		return &r.State, nil
	}
	return mesos_v1.TaskState_TASK_UNKNOWN.Enum(), nil
}

// AllByState gets all tasks that match the given state.
func (m *SprintTaskHandler) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}
	tasks := []*mesos_v1.TaskInfo{}
	for _, v := range m.tasks {
		if v.State == state {
			tasks = append(tasks, v.Info)
		}
	}

	if len(tasks) == 0 {
		return nil, errors.New("No tasks with state " + state.String())
	}

	return tasks, nil
}

// All gets all tasks that are known to the task manager.
func (m *SprintTaskHandler) All() ([]manager.Task, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}
	allTasks := []manager.Task{}
	for _, task := range m.tasks {
		allTasks = append(allTasks, task)
	}
	return allTasks, nil
}

// Function that wraps writing to the storage backend.
func (m *SprintTaskHandler) storageWrite(id string, encoded *bytes.Buffer) func() error {
	return func() error {
		err := m.storage.Update(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes()))
		if err != nil {
			m.logger.Emit(
				logging.ERROR, "Failed to update task %s with name %s to persistent data store. Retrying...",
				id,
			)
		}
		return err
	}
}

// Function that wraps deleting from the storage backend.
func (m *SprintTaskHandler) storageDelete(taskId string) func() error {
	return func() error {
		err := m.storage.Delete(TASK_DIRECTORY + taskId)
		if err != nil {
			m.logger.Emit(logging.ERROR, err.Error())
		}
		return err
	}
}

// Encodes task data to a small, efficient payload that can be transmitted across the wire.
func (m *SprintTaskHandler) encode(task *mesos_v1.TaskInfo, state mesos_v1.TaskState) error {
	// Panics on nil values.
	err := m.encoder.Encode(manager.Task{
		Info:  task,
		State: state,
	})

	return err
}

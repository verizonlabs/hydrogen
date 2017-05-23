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
	"time"
)

//
//Satisfies the TaskManager interface in sdk
//
//Sprint Task Manager integrates logging and storage backend.
//All tasks are written only during creation, updates and deletes.
//Reads are reserved for reconciliation calls.
//

const (
	TASK_DIRECTORY = "/tasks/"
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
		tasks   structures.DistributedMap
		storage persistence.Storage
		retries structures.DistributedMap
		config  *scheduler.Configuration
		logger  logging.Logger
	}
)

// Returns the core task manager that's used by the scheduler.
func NewTaskManager(
	cmap structures.DistributedMap,
	storage persistence.Storage,
	config *scheduler.Configuration,
	logger logging.Logger) SprintTaskManager {

	return &SprintTaskHandler{
		tasks:   cmap,
		storage: storage,
		retries: structures.NewConcurrentMap(),
		config:  config,
		logger:  logger,
	}
}

// Encodes task data to a small, efficient payload that can be transmitted across the wire.
func (m *SprintTaskHandler) encode(task *mesos_v1.TaskInfo, state mesos_v1.TaskState) (bytes.Buffer, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	// Panics on nil values.
	err := e.Encode(manager.Task{
		Info:  task,
		State: state,
	})
	if err != nil {
		return b, err
	}

	return b, nil
}

//
// Methods to satisfy retry interface.
//
func (s *SprintTaskHandler) AddPolicy(policy *task.TimeRetry, mesosTask *mesos_v1.TaskInfo) error {
	if mesosTask != nil {
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
func (s *SprintTaskHandler) CheckPolicy(mesosTask *mesos_v1.TaskInfo) (*retry.TaskRetry, error) {
	if mesosTask != nil {
		policy := s.retries.Get(mesosTask.GetName())
		if policy != nil {
			return policy.(*retry.TaskRetry), nil
		}
	}

	return nil, errors.New("No policy exists for this task.")
}

// Removes an existing policy associated with the given task.
func (s *SprintTaskHandler) ClearPolicy(mesosTask *mesos_v1.TaskInfo) error {
	if mesosTask != nil {
		s.retries.Delete(mesosTask.GetTaskId().GetValue())
		return nil
	}
	return errors.New("Nil mesos task passed in")
}

//
// End of Methods to satisfy Retry Interface
//

//
// Task Manager Methods
//

// Adds and persists a new task into the task manager.
// Duplicate task names are not allowed by Mesos, thus they are not allowed here.
func (m *SprintTaskHandler) Add(t *mesos_v1.TaskInfo) error {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Emit(logging.INFO, "Recovered in Task manager Add", r)
			return
		}
	}()

	name := t.GetName()
	if m.tasks.Get(name) != nil {
		return errors.New("Task " + name + " already exists")
	}

	// Write forward.
	encoded, err := m.encode(t, manager.UNKNOWN)
	if err != nil {
		return err
	}

	id := t.TaskId.GetValue()
	policy, _ := m.storage.CheckPolicy(nil)
	err = m.storage.RunPolicy(policy, func() error {
		err := m.storage.Create(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes()))
		if err != nil {
			m.logger.Emit(
				logging.ERROR,
				"Failed to save task %s with name %s to persistent data store. Retrying...",
				id,
				t.GetName(),
			)
		}

		return err
	})

	if err != nil {
		return err
	}

	m.tasks.Set(t.GetName(), manager.Task{
		State: manager.UNKNOWN,
		Info:  t,
	})

	return nil
}

// Deletes a task from memory and etcd, and clears any associated policy.
func (m *SprintTaskHandler) Delete(task *mesos_v1.TaskInfo) error {
	policy, _ := m.storage.CheckPolicy(nil)
	err := m.storage.RunPolicy(policy, func() error {
		err := m.storage.Delete(TASK_DIRECTORY + task.GetTaskId().GetValue())
		if err != nil {
			m.logger.Emit(logging.ERROR, err.Error())
		}

		return err
	})

	if err != nil {
		return err
	}

	m.tasks.Delete(task.GetName())
	m.ClearPolicy(task)

	return nil
}

// Gets a task by its name.
func (m *SprintTaskHandler) Get(name *string) (*mesos_v1.TaskInfo, error) {
	ret := m.tasks.Get(*name)
	if ret != nil {
		return ret.(manager.Task).Info, nil
	}

	return nil, errors.New("Could not find task.")
}

// Gets a task by its ID.
func (m *SprintTaskHandler) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	if m.tasks.Length() == 0 {
		return nil, errors.New("Task manager is empty.")
	}

	for v := range m.tasks.Iterate() {
		t := v.Value.(manager.Task)
		if t.Info.GetTaskId().GetValue() == id.GetValue() {
			return t.Info, nil
		}
	}

	return nil, errors.New("Could not find task by id: " + id.GetValue())
}

// Indicates whether or not the task manager holds the specified task.
func (m *SprintTaskHandler) HasTask(task *mesos_v1.TaskInfo) bool {
	ret := m.tasks.Get(task.GetName())
	if ret == nil {
		return false
	}

	return true
}

// Returns the total number of tasks that the task manager holds.
func (m *SprintTaskHandler) TotalTasks() int {
	return m.tasks.Length()
}

// Returns the internal structure used to hold task data.
func (m *SprintTaskHandler) Tasks() structures.DistributedMap {
	return m.tasks
}

// Update the given task with the given state.
func (m *SprintTaskHandler) Set(state mesos_v1.TaskState, t *mesos_v1.TaskInfo) error {
	// Write forward.
	encoded, err := m.encode(t, state)
	if err != nil {
		m.logger.Emit(logging.INFO, err.Error())
	}

	id := t.TaskId.GetValue()

	policy, _ := m.storage.CheckPolicy(nil)
	err = m.storage.RunPolicy(policy, func() error {
		err := m.storage.Update(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes()))
		if err != nil {
			m.logger.Emit(
				logging.ERROR, "Failed to update task %s with name %s to persistent data store. Retrying...",
				id,
				t.GetName(),
			)
		}

		return err
	})

	if err != nil {
		return err
	}

	m.tasks.Set(t.GetName(), manager.Task{
		Info:  t,
		State: state,
	})

	return nil
}

// Gets the task's state based on the supplied task name.
func (m *SprintTaskHandler) State(name *string) (*mesos_v1.TaskState, error) {
	ret := m.tasks.Get(*name)
	if ret != nil {
		state := ret.(manager.Task).State
		return &state, nil
	}

	return nil, errors.New("Task " + *name + " does not exist")
}

// Gets all tasks that match the given state.
func (m *SprintTaskHandler) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	tasks := []*mesos_v1.TaskInfo{}
	for v := range m.tasks.Iterate() {
		t := v.Value.(manager.Task)
		if t.State == state {
			tasks = append(tasks, t.Info)
		}
	}

	if len(tasks) == 0 {
		return nil, errors.New("No tasks found with state of " + state.String())
	}

	return tasks, nil
}

// Gets all tasks that are known to the task manager.
func (m *SprintTaskHandler) All() ([]manager.Task, error) {
	if m.tasks.Length() == 0 {
		return nil, errors.New("Task manager is empty.")
	}

	tasks := []manager.Task{}
	for v := range m.tasks.Iterate() {
		tasks = append(tasks, v.Value.(manager.Task))
	}

	return tasks, nil
}

//
// End of task manager methods.
//

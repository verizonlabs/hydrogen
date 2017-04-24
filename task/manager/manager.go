package manager

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"os"
	"sprint/scheduler"
	"sprint/task/retry"
	"time"
)

/*
Satisfies the TaskManager interface in sdk

Sprint Task Manager integrates logging and storage backend.
All tasks are written only during creation, updates and deletes.
Reads are reserved for reconciliation calls.


NOTE (tim): We store tasks with a particular set of metadata in the struct
"Task".  However the interface specifies that we get a task info back,
rather than our custom struct back.

We need to make the storage of our tasks consistent, as well as be able to handle
additional metadata about tasks.

*/

var IS_TESTING bool

// NOTE (tim): Put this in the utils package or somewhere else?
func IsTesting() bool {
	if os.Getenv("TESTING") == "true" {
		return true
	}
	return false
}

const (
	TASK_DIRECTORY = "/tasks/"
)

type (
	SprintTaskManager interface {
		manager.TaskManager
		retry.Retry
	}
	SprintTaskHandler struct {
		tasks   structures.DistributedMap
		storage etcd.KeyValueStore
		retries map[string]*retry.TaskRetry
		config  *scheduler.Configuration
		logger  logging.Logger
	}
)

func NewTaskManager(
	cmap structures.DistributedMap,
	storage etcd.KeyValueStore,
	config *scheduler.Configuration,
	logger logging.Logger) SprintTaskManager {

	IS_TESTING = IsTesting()
	return &SprintTaskHandler{
		tasks:   cmap,
		storage: storage,
		retries: make(map[string]*retry.TaskRetry),
		config:  config,
		logger:  logger,
	}
}

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

		s.retries[mesosTask.GetName()] = &retry.TaskRetry{
			TotalRetries: 0,
			RetryTime:    t,
			Backoff:      policy.Backoff,
			Name:         mesosTask.GetName(),
		}
		return nil
	}
	return errors.New("Nil mesos task passed in")
}

func (s *SprintTaskHandler) RunPolicy(policy *retry.TaskRetry, f func()) error {
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
	time.AfterFunc(delay, f)
	return nil
}

func (s *SprintTaskHandler) CheckPolicy(mesosTask *mesos_v1.TaskInfo) (*retry.TaskRetry, error) {
	if mesosTask != nil {
		if policy, ok := s.retries[mesosTask.GetName()]; ok {
			return policy, nil
		}
	}
	return nil, errors.New("No policy exists for this task.")
}

func (s *SprintTaskHandler) ClearPolicy(mesosTask *mesos_v1.TaskInfo) error {
	if mesosTask != nil {
		delete(s.retries, mesosTask.GetTaskId().GetValue())
		return nil
	}
	return errors.New("Nil mesos task passed in")
}

///

func (m *SprintTaskHandler) Add(t *mesos_v1.TaskInfo) error {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Emit(logging.INFO, "Recovered in Task manager Add", r)
			return
		}
	}()
	// Write forward.
	encoded, err := m.encode(t, manager.UNKNOWN)
	if err != nil {
		return err
	}
	id := t.TaskId.GetValue()

	for {
		if err := m.storage.Create(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes())); err != nil {
			if IS_TESTING {
				return errors.New("Failed to ADD.")
			}

			m.logger.Emit(logging.ERROR, "Failed to save task %s with name %s to persistent data store", id, t.GetName())
			time.Sleep(m.config.Persistence.RetryInterval)
			continue
		}
		break
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

func (m *SprintTaskHandler) Delete(task *mesos_v1.TaskInfo) {
	for {
		err := m.storage.Delete(TASK_DIRECTORY + task.GetTaskId().GetValue())
		if err != nil {
			if IS_TESTING {
				return
			}

			m.logger.Emit(logging.ERROR, err.Error())
			time.Sleep(m.config.Persistence.RetryInterval)
			continue
		}
		break
	}

	m.tasks.Delete(task.GetName())
	m.ClearPolicy(task)
}

func (m *SprintTaskHandler) Get(name *string) (*mesos_v1.TaskInfo, error) {
	ret := m.tasks.Get(*name)
	if ret != nil {
		return ret.(manager.Task).Info, nil
	}

	return nil, errors.New("Could not find task.")
}

// Check to see if any tasks we have match the id passed in.
func (m *SprintTaskHandler) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
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

func (m *SprintTaskHandler) HasTask(task *mesos_v1.TaskInfo) bool {
	ret := m.tasks.Get(task.GetName())
	if ret == nil {
		return false
	}

	return true
}

func (m *SprintTaskHandler) TotalTasks() int {
	return m.tasks.Length()
}

func (m *SprintTaskHandler) Tasks() structures.DistributedMap {
	return m.tasks
}

// Update a task with a certain state.
func (m *SprintTaskHandler) Set(state mesos_v1.TaskState, t *mesos_v1.TaskInfo) {
	// Write forward.
	encoded, err := m.encode(t, state)
	if err != nil {
		m.logger.Emit(logging.INFO, err.Error())
	}

	id := t.TaskId.GetValue()

	for {
		if err := m.storage.Update(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes())); err != nil {
			if IS_TESTING {
				return
			}
			m.logger.Emit(logging.ERROR, "Failed to update task %s with name %s to persistent data store", id, t.GetName())
			// NOTE(tim): This will hang everything if we retry forever.
			time.Sleep(m.config.Persistence.RetryInterval)
			continue
		}
		break
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
func (m *SprintTaskHandler) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
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

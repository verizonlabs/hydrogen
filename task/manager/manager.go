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
	TASK_DIRECTORY    = "/tasks/"
	DELETE         OP = "delete"
	ADD            OP = "add"
	GET            OP = "get"
	GETBYID        OP = "getbyid"
	HASTASK        OP = "hastask"
	SET            OP = "set"
	ALLBYSTATE     OP = "allbystate"
	ALL            OP = "all"
	STATE          OP = "state"
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
		tasks      map[string]manager.Task
		storage    persistence.Storage
		retries    structures.DistributedMap
		config     *scheduler.Configuration
		task       chan *mesos_v1.TaskInfo
		logger     logging.Logger
		writeQueue chan WriteResponse
		readQueue  chan ReadResponse
	}

	WriteResponse struct {
		task  *mesos_v1.TaskInfo
		state mesos_v1.TaskState
		reply chan error
		op    OP
	}

	ReadResponse struct {
		name       string
		id         string
		state      mesos_v1.TaskState
		reply      chan *mesos_v1.TaskInfo
		replyState chan mesos_v1.TaskState
		replyTask  chan manager.Task
		op         OP
	}
	OP string
)

// Returns the core task manager that's used by the scheduler.
func NewTaskManager(
	cmap map[string]manager.Task,
	storage persistence.Storage,
	config *scheduler.Configuration,
	logger logging.Logger) SprintTaskManager {

	handler := &SprintTaskHandler{
		tasks:      cmap,
		storage:    storage,
		retries:    structures.NewConcurrentMap(),
		config:     config,
		task:       make(chan *mesos_v1.TaskInfo),
		logger:     logger,
		writeQueue: make(chan WriteResponse, 100),
		readQueue:  make(chan ReadResponse, 100),
	}
	go handler.listen()
	return handler
}

func (m *SprintTaskHandler) listen() {
	for {
		select {
		case write := <-m.writeQueue:
			switch write.op {
			case ADD:
				m.add(write)
			case DELETE:
				m.delete(write)
			case SET:
				m.set(write)
			}
		case read := <-m.readQueue:
			switch read.op {
			case GET:
				m.get(read)
			case GETBYID:
				m.getById(read)
			case HASTASK:
				m.hasTask(read)
			case STATE:
				m.state(read)
			case ALL:
				m.all(read)
			case ALLBYSTATE:
				m.allByState(read)
			}
		default:
		}
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
	// Add a request to the AddQueue.
	// Make a response struct to pass a channel to listen
	// for a response.
	reply := make(chan error, 1)
	r := WriteResponse{task: t, reply: reply, op: ADD}
	m.writeQueue <- r
	response := <-r.reply // block until reply occurs.
	if response != nil {
		return response
	}
	return response
}

func (m *SprintTaskHandler) add(add WriteResponse) {
	task := add.task

	// Use a unique ID for storing in the map, taskid?
	if _, ok := m.tasks[task.GetName()]; ok {
		add.reply <- errors.New("Task " + task.GetName() + " already exists")
	}

	// Write forward.
	encoded, err := m.encode(task, manager.UNKNOWN)
	if err != nil {
		add.reply <- err
	}

	id := task.TaskId.GetValue()

	// TODO (tim): Policy should be set once for all storage operations and taken care of by storage class.
	// Storage.Create() is called and it's default policy is run.

	policy, _ := m.storage.CheckPolicy(nil)

	err = m.storage.RunPolicy(policy, func() error {
		err := m.storage.Create(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes()))
		if err != nil {
			m.logger.Emit(
				logging.ERROR,
				"Failed to save task %s with name %s to persistent data store. Retrying...",
				id,
				task.GetName(),
			)
		}

		return err
	})

	if err != nil {
		add.reply <- err
	}

	m.tasks[task.GetName()] = manager.Task{
		State: manager.UNKNOWN,
		Info:  task,
	}

	add.reply <- nil
}

// Deletes a task from memory and etcd, and clears any associated policy.
func (m *SprintTaskHandler) Delete(task *mesos_v1.TaskInfo) error {
	reply := make(chan error)
	r := WriteResponse{task: task, reply: reply, op: "delete"}
	m.writeQueue <- r
	response := <-r.reply // block until reply occurs.
	close(r.reply)
	if response != nil {
		return response
	}
	return response
}

func (m *SprintTaskHandler) delete(res WriteResponse) {
	task := res.task
	policy, _ := m.storage.CheckPolicy(nil)
	err := m.storage.RunPolicy(policy, func() error {
		err := m.storage.Delete(TASK_DIRECTORY + task.GetTaskId().GetValue())
		if err != nil {
			m.logger.Emit(logging.ERROR, err.Error())
		}

		return err
	})

	if err != nil {
		res.reply <- err
	}
	delete(m.tasks, task.GetName())
	m.ClearPolicy(task)

	res.reply <- nil
}

// Gets a task by its name.
func (m *SprintTaskHandler) Get(name *string) (*mesos_v1.TaskInfo, error) {
	reply := make(chan *mesos_v1.TaskInfo, 1)
	r := ReadResponse{name: *name, reply: reply, op: GET}
	m.readQueue <- r
	response := <-r.reply // block until reply occurs.
	if response == nil {
		return nil, errors.New("Could not find task.")
	}
	return response, nil
}

func (m *SprintTaskHandler) get(res ReadResponse) {
	name := res.name
	ret := m.tasks[name]
	res.reply <- ret.Info
}

// Gets a task by its ID.
func (m *SprintTaskHandler) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty.")
	}
	reply := make(chan *mesos_v1.TaskInfo)
	r := ReadResponse{id: id.GetValue(), reply: reply, op: GETBYID}
	m.readQueue <- r
	response := <-r.reply // block until reply occurs.
	// If channel returns nil, nothing was found.
	if response.TaskId == nil {
		return nil, errors.New("Could not find task by id: " + id.GetValue())
	}
	// If channel wasn't nil, no error.
	return response, nil
}

func (m *SprintTaskHandler) getById(ret ReadResponse) {
	id := ret.id
	for _, v := range m.tasks {
		if id == v.Info.GetTaskId().GetValue() {
			ret.reply <- v.Info
			return
		}
	}
	ret.reply <- &mesos_v1.TaskInfo{}
}

// Indicates whether or not the task manager holds the specified task.
func (m *SprintTaskHandler) HasTask(task *mesos_v1.TaskInfo) bool {
	if len(m.tasks) == 0 {
		return false // If the manager is empty, this one doesn't exist.
	}
	reply := make(chan manager.Task, 1)
	r := ReadResponse{name: task.GetName(), replyTask: reply, op: HASTASK}
	m.readQueue <- r
	response := <-r.replyTask // block until reply occurs.
	// If channel returns nil, nothing was found.
	if response.Info == nil {
		return false
	}
	// If channel wasn't nil, no error.
	return true
}

func (m *SprintTaskHandler) hasTask(ret ReadResponse) {
	if val, ok := m.tasks[ret.name]; ok {
		ret.replyTask <- val
	} else {
		ret.replyTask <- manager.Task{}
	}
}

// Returns the total number of tasks that the task manager holds.
func (m *SprintTaskHandler) TotalTasks() int {
	return len(m.tasks)
}

// Update the given task with the given state.
func (m *SprintTaskHandler) Set(state mesos_v1.TaskState, t *mesos_v1.TaskInfo) error {
	reply := make(chan error, 1)
	r := WriteResponse{task: t, state: state, reply: reply, op: SET}
	m.writeQueue <- r
	response := <-r.reply
	if response != nil {
		return errors.New("Failed to set state on task")
	}
	return nil
}

func (m *SprintTaskHandler) set(ret WriteResponse) {
	// Write forward.
	encoded, err := m.encode(ret.task, ret.state)
	if err != nil {
		m.logger.Emit(logging.INFO, err.Error())
	}

	id := ret.task.TaskId.GetValue()

	policy, _ := m.storage.CheckPolicy(nil)
	err = m.storage.RunPolicy(policy, func() error {
		err := m.storage.Update(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes()))
		if err != nil {
			m.logger.Emit(
				logging.ERROR, "Failed to update task %s with name %s to persistent data store. Retrying...",
				id,
				ret.task.GetName(),
			)
		}

		return err
	})

	if err != nil {
		ret.reply <- err
	}

	m.tasks[ret.task.GetName()] = manager.Task{
		Info:  ret.task,
		State: ret.state,
	}
	ret.reply <- nil
}

// Gets the task's state based on the supplied task name.
func (m *SprintTaskHandler) State(name *string) (*mesos_v1.TaskState, error) {
	reply := make(chan mesos_v1.TaskState, 1)
	r := ReadResponse{name: *name, replyState: reply, op: STATE}
	m.readQueue <- r
	response := <-r.replyState // block until reply occurs.
	return &response, nil
}

func (m *SprintTaskHandler) state(ret ReadResponse) {
	if r, ok := m.tasks[ret.name]; ok {
		ret.replyState <- r.State
	} else {
		ret.replyState <- mesos_v1.TaskState_TASK_UNKNOWN
	}
}

// Gets all tasks that match the given state.
func (m *SprintTaskHandler) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}

	reply := make(chan *mesos_v1.TaskInfo)
	r := ReadResponse{state: state, reply: reply, op: ALLBYSTATE}
	m.readQueue <- r

	tasks := make([]*mesos_v1.TaskInfo, 0)
	for task := range r.reply {
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		return nil, errors.New("No tasks with state " + state.String())
	}

	return tasks, nil
}

func (m *SprintTaskHandler) allByState(read ReadResponse) {
	for _, v := range m.tasks {
		if v.State == read.state {
			read.reply <- v.Info
		}
	}
	close(read.reply) // let the calling class know we're done sending values.
}

// Gets all tasks that are known to the task manager.
func (m *SprintTaskHandler) All() ([]manager.Task, error) {
	if len(m.tasks) == 0 {
		return nil, errors.New("Task manager is empty")
	}

	reply := make(chan manager.Task)
	r := ReadResponse{replyTask: reply, op: ALL}
	m.readQueue <- r
	tasks := []manager.Task{}
	for task := range r.replyTask {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *SprintTaskHandler) all(read ReadResponse) {
	for _, v := range m.tasks {
		read.replyTask <- v
	}
	close(read.replyTask)
}

//
// End of task manager methods.
//

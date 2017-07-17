package manager

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"sprint/scheduler"
	"sprint/task/persistence"
	"sprint/task/retry"
	"strings"
	"time"
	"bytes"
	"encoding/gob"
)

const (
	GROUP_SIZE         = "/size/"
	GROUP_DIRECTORY    = "/taskgroup/"
	TASK_DIRECTORY     = "/tasks/"
	UNLINK          OP = 0
	LINK            OP = 1
	DELETE          OP = 2
	ADD             OP = 3
	GET             OP = 4
	GETBYID         OP = 5
	HASTASK         OP = 6
	SET             OP = 7
	ALLBYSTATE      OP = 8
	ALL             OP = 9
	STATE           OP = 10
	LEN             OP = 11
	CREATEGROUP     OP = 12
	DELGROUP        OP = 13
	SETGROUPSIZE    OP = 14
	READGROUP       OP = 15
	ISINGROUP       OP = 16
)

type (
	// Provides pluggable task managers that the framework can work with.
	// Also used extensively for testing with mocks.
	SprintTaskManager interface {
		manager.TaskManager
		retry.Retry
		ScaleGrouping
	}

	// Our primary task handler that implements the above interface.
	// The task handler manages all tasks that are submitted, updated, or deleted.
	// Offers from Mesos are matched up with user-submitted tasks, and those tasks are updated via event callbacks.
	SprintTaskHandler struct {
		buffer     bytes.Buffer
		encoder    *gob.Encoder
		tasks      map[string]manager.Task
		groups     map[string][]*mesos_v1.AgentID
		storage    persistence.Storage
		retries    structures.DistributedMap
		config     *scheduler.Configuration
		task       chan *mesos_v1.TaskInfo
		logger     logging.Logger
		writeQueue chan Write
		readQueue  chan Read
	}

	// Operation to switch on
	OP uint8

	// Send a write request
	Write struct {
		task    *mesos_v1.TaskInfo
		group   string
		agentID *mesos_v1.AgentID
		size    int
		state   mesos_v1.TaskState
		reply   chan error
		op      OP
	}

	// Send a read request
	Read struct {
		name       string
		id         string
		state      mesos_v1.TaskState
		taskInfo   *mesos_v1.TaskInfo
		agents     chan []*mesos_v1.AgentID
		isInGroup  chan bool
		reply      chan *mesos_v1.TaskInfo
		replyState chan mesos_v1.TaskState
		replyTask  chan manager.Task
		replySize  chan int
		op         OP
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

	var b bytes.Buffer
	handler := &SprintTaskHandler{
		buffer:     b,
		encoder:    gob.NewEncoder(&b),
		tasks:      cmap,
		storage:    storage,
		retries:    structures.NewConcurrentMap(),
		config:     config,
		task:       make(chan *mesos_v1.TaskInfo),
		groups:     make(map[string][]*mesos_v1.AgentID),
		logger:     logger,
		writeQueue: make(chan Write, 100),
		readQueue:  make(chan Read, 100),
	}
	go handler.listen()
	return handler
}

// Method that listens for r/w channel events.
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
			case SETGROUPSIZE:
				m.setSize(write)
			case DELGROUP:
				m.deleteGroup(write)
			case CREATEGROUP:
				m.createGroup(write)
			case UNLINK:
				m.unlink(write)
			case LINK:
				m.link(write)
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
			case LEN:
				m.totalTasks(read)
			case READGROUP:
				m.readGroup(read)
			case ISINGROUP:
				m.isInGroup(read)
			}
		}
	}
}

// Checks if a task is in a grouping.
func (m *SprintTaskHandler) IsInGroup(task *mesos_v1.TaskInfo) bool {
	ch := make(chan bool, 1)
	strippedName := strings.Split(task.GetName(), "-")[0]
	m.readQueue <- Read{isInGroup: ch, name: strippedName, op: ISINGROUP}
	response := <-ch
	return response
}

// Creates a group for the given name.
func (m *SprintTaskHandler) CreateGroup(name string) error {
	ch := make(chan error, 1)
	w := Write{reply: ch, group: name, op: CREATEGROUP}
	m.writeQueue <- w
	response := <-w.reply
	return response
}

// Increment or decrement a group size by the given amount on the group given name.
func (m *SprintTaskHandler) SetSize(name string, amount int) error {
	ch := make(chan error)
	strippedName := strings.Split(name, "-")[0]
	m.writeQueue <- Write{reply: ch, group: strippedName, size: amount, op: SETGROUPSIZE}
	response := <-ch
	return response
}

// Read the agents tied to a grouping.
func (m *SprintTaskHandler) ReadGroup(name string) []*mesos_v1.AgentID {
	ch := make(chan []*mesos_v1.AgentID)
	strippedName := strings.Split(name, "-")[0]
	m.readQueue <- Read{agents: ch, name: strippedName, op: READGROUP}
	response := <-ch
	return response
}

// Link an agentID to a group name.
func (m *SprintTaskHandler) Link(name string, agent *mesos_v1.AgentID) error {
	ch := make(chan error)
	strippedName := strings.Split(name, "-")[0]
	m.writeQueue <- Write{reply: ch, group: strippedName, agentID: agent, op: LINK}
	response := <-ch
	return response
}

// Unlink removes an task/agent mapping.
func (m *SprintTaskHandler) Unlink(name string, agent *mesos_v1.AgentID) error {
	ch := make(chan error)
	strippedName := strings.Split(name, "-")[0]
	m.writeQueue <- Write{reply: ch, group: strippedName, agentID: agent, op: UNLINK}
	response := <-ch
	return response
}

func (m *SprintTaskHandler) DeleteGroup(name string) error {
	ch := make(chan error)
	strippedName := strings.Split(name, "-")[0]
	m.writeQueue <- Write{reply: ch, group: strippedName, op: DELGROUP}
	response := <-ch
	return response
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

// Adds and persists a new task into the task manager.
// Duplicate task names are not allowed by Mesos, thus they are not allowed here.
func (m *SprintTaskHandler) Add(t *mesos_v1.TaskInfo) error {
	reply := make(chan error, 1)
	r := Write{task: t, reply: reply, op: ADD}
	m.writeQueue <- r
	response := <-r.reply

	return response
}

// Deletes a task from memory and etcd, and clears any associated policy.
func (m *SprintTaskHandler) Delete(task *mesos_v1.TaskInfo) error {
	reply := make(chan error)
	r := Write{task: task, reply: reply, op: DELETE}
	m.writeQueue <- r
	response := <-r.reply
	close(r.reply)
	return response
}

// Gets a task by its name.
func (m *SprintTaskHandler) Get(name *string) (*mesos_v1.TaskInfo, error) {
	reply := make(chan *mesos_v1.TaskInfo, 1)
	r := Read{name: *name, reply: reply, op: GET}
	m.readQueue <- r
	response := <-r.reply
	if response == nil {
		return nil, errors.New("Could not find task.")
	}
	return response, nil
}

// Gets a task by its ID.
func (m *SprintTaskHandler) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	if m.TotalTasks() == 0 {
		return nil, errors.New("Task manager is empty.")
	}
	reply := make(chan *mesos_v1.TaskInfo)
	r := Read{id: id.GetValue(), reply: reply, op: GETBYID}
	m.readQueue <- r
	response := <-r.reply
	if response == nil {
		return nil, errors.New("Could not find task by id: " + id.GetValue())
	}
	return response, nil
}

// Indicates whether or not the task manager holds the specified task.
func (m *SprintTaskHandler) HasTask(task *mesos_v1.TaskInfo) bool {
	if m.TotalTasks() == 0 {
		return false // If the manager is empty, this one doesn't exist.
	}
	reply := make(chan manager.Task, 1)
	r := Read{name: task.GetName(), replyTask: reply, op: HASTASK}
	m.readQueue <- r
	response := <-r.replyTask

	return response.Info != nil
}

// Returns the total number of tasks that the task manager holds.
func (m *SprintTaskHandler) TotalTasks() int {
	replyLength := make(chan int, 1)
	r := Read{replySize: replyLength, op: LEN}
	m.readQueue <- r
	response := <-r.replySize
	return response
}

// Update the given task with the given state.
func (m *SprintTaskHandler) Set(state mesos_v1.TaskState, t *mesos_v1.TaskInfo) error {
	reply := make(chan error, 1)
	r := Write{task: t, state: state, reply: reply, op: SET}
	m.writeQueue <- r
	response := <-r.reply
	if response != nil {
		return errors.New("Failed to set state on task")
	}
	return nil
}

// Gets the task's state based on the supplied task name.
func (m *SprintTaskHandler) State(name *string) (*mesos_v1.TaskState, error) {
	reply := make(chan mesos_v1.TaskState, 1)
	r := Read{name: *name, replyState: reply, op: STATE}
	m.readQueue <- r
	response := <-r.replyState
	return &response, nil
}

// Gets all tasks that match the given state.
func (m *SprintTaskHandler) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	if m.TotalTasks() == 0 {
		return nil, errors.New("Task manager is empty")
	}

	reply := make(chan *mesos_v1.TaskInfo)
	r := Read{state: state, reply: reply, op: ALLBYSTATE}

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

// Gets all tasks that are known to the task manager.
func (m *SprintTaskHandler) All() ([]manager.Task, error) {
	if m.TotalTasks() == 0 {
		return nil, errors.New("Task manager is empty")
	}
	reply := make(chan manager.Task)
	r := Read{replyTask: reply, op: ALL}
	m.readQueue <- r
	tasks := []manager.Task{}
	for task := range r.replyTask {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

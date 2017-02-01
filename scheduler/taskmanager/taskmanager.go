package taskmanager

import (
	"github.com/orcaman/concurrent-map"
	"mesos-sdk"
	"mesos-sdk/taskmngr"
	"stash.verizon.com/dkt/mlog"
)

type TaskState struct {
	taskinfo mesos.TaskInfo
	id       int
	status   string
}

// Satisfy the Task interface
func (t *TaskState) Info() (mesos.TaskInfo, error) {
	return t.taskinfo, nil
}

func (t *TaskState) SetInfo(taskinfo mesos.TaskInfo) {
	t.taskinfo = taskinfo
}

func (t *TaskState) Id() (int, error) {
	return t.id, nil
}

func (t *TaskState) SetId(id int) error {
	t.id = id
	return nil
}

func (t *TaskState) Status() (string, error) {
	return t.status, nil
}

func (t *TaskState) SetStatus(status string) error {
	t.status = status
	return nil
}

type Manager struct {
	frameworkId string
	totalTasks  int
	tasks       *cmap.ConcurrentMap
}

func NewManager(conmap *cmap.ConcurrentMap) *Manager {
	return &Manager{tasks: conmap}
}

// Satisfy the TaskManager interface

// Provision a task
func (m *Manager) Provision(task taskmngr.Task) error {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return err
	}
	m.tasks.Set(string(id), task)
	return nil
}

// Delete a task
func (m *Manager) Delete(task taskmngr.Task) error {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return err
	}
	m.tasks.Remove(string(id))
	return nil
}

// Set a task status
func (m *Manager) SetTask(task taskmngr.Task, status string) error {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return err
	}
	task.SetStatus(status)
	m.tasks.Set(string(id), task) // Overwrite the old value here.
	return nil
}

// Check if a task has a particular status.
func (m *Manager) IsTask(task taskmngr.Task, status string) (bool, error) {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return false, err
	}
	t, exists := m.tasks.Get(string(id))
	if !exists {
		return false, nil
	}

	currentStatus, err := t.(taskmngr.Task).Status()
	if err != nil {
		mlog.Error(err.Error())
		return false, nil
	}

	return currentStatus == status, nil
}

// Check if the task is already in the task manager.
func (m *Manager) HasTask(task taskmngr.Task) (bool, error) {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return false, err
	}
	return m.tasks.Has(string(id)), nil
}

// Check if we have tasks left to execute.
func (m *Manager) HasQueuedTasked() (bool, error) {
	return !m.tasks.IsEmpty(), nil
}

func (m *Manager) Tasks() *cmap.ConcurrentMap {
	return m.tasks
}

// Create a map of taskid->agentid
func (m *Manager) TaskIdAgentIdMap(task taskmngr.Task) (map[string]string, error) {
	ret := make(map[string]string, m.Tasks().Count())
	if ok, _ := m.HasQueuedTasked(); ok {
		for task := range m.tasks.IterBuffered() {
			t := task.Val.(mesos.TaskInfo)
			ret[t.TaskID.Value] = t.AgentID.Value
		}
	}
	return ret, nil
}

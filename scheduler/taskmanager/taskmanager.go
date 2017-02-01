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

func (t *TaskState) Status() error {
	return t.status
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
	m.tasks.Set(task.Id(), task)
	return nil
}

// Delete a task
func (m *Manager) Delete(task taskmngr.Task) error {
	delete(m.tasks, task.Id())
	return nil
}

// Set a task status
func (m *Manager) SetTask(task taskmngr.Task, status string) error {
	task.SetStatus(status)
	m.tasks.Set(task.Id(), task) // Overwrite the old value here.
}

// Check if a task has a particular status.
func (m *Manager) IsTask(task taskmngr.Task, status string) bool, error {
	task, err := m.tasks.Get(task.Id())
	if err != nil{
		mlog.Error(err)
		return false, err
	}
	return task.(taskmngr.Task).Status() == status, nil
}

// Check if the task is already in the task manager.
func (m *Manager) HasTask(task taskmngr.Task) (bool, error){
	return m.tasks.Has(task.Id()), nil
}

// Check if we have tasks left to execute.
func (m *Manager) HasQueuedTasked()(bool, error){
	return !m.tasks.IsEmpty()
}

package taskmanager

import (
	"mesos-sdk"
	"stash.verizon.com/dkt/mlog"
)

type TaskState struct {
	taskinfo mesos.Task
	id       string
	status   string
}

// Satisfy the Task interface
func (t *TaskState) Info() (mesos.Task, error) {
	return t.taskinfo, nil
}

func (t *TaskState) SetInfo(taskinfo mesos.Task) error {
	t.taskinfo = taskinfo
	return nil
}

func (t *TaskState) Id() (string, error) {
	return t.id, nil
}

func (t *TaskState) SetId(id string) error {
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
	tasks       map[string]mesos.Task
}

func NewManager() *Manager {
	return &Manager{tasks: make(map[string]mesos.Task)}
}

// Satisfy the TaskManager interface

// Provision a task
func (m *Manager) Add(task mesos.Task) error {
	m.tasks[task.GetTaskId().Value] = task

	return nil
}

// Delete a task
func (m *Manager) Delete(id string) {
	delete(m.tasks, id)
}

// Set a task status
func (m *Manager) SetTask(task mesos.Task, status mesos.TaskState) error {
	task.State = status
	m.tasks[task.TaskId] = task // Overwrite the old value here.
	return nil
}

// Check if a task has a particular status.
func (m *Manager) IsTask(task mesos.Task, status string) (bool, error) {
	if _, ok := m.tasks[task.GetTaskId().Value]; !ok {
		return false, nil
	}
	return task.State == status, nil
}

// Check if the task is already in the task manager.
func (m *Manager) HasTask(task mesos.Task) (bool, error) {
	if _, ok := m.tasks[task.GetTaskId().GetValue()]; ok {
		return false, nil
	}
	return true, nil
}

// Check if we have tasks left to execute.
func (m *Manager) HasQueuedTasks() (bool, error) {
	mlog.Error("has any tasks?", len(m.tasks))
	return !(len(m.tasks) == 0), nil
}

func (m *Manager) Tasks() map[string]mesos.Task {
	return m.tasks
}

// Create a map of taskid->agentid
func TaskIdAgentIdMap(m map[string]mesos.Task) (map[string]string, error) {
	ret := make(map[string]string, len(m))
	for k, v := range m {
		taskInfo := v.GetAgentId().Value
		ret[k] = taskInfo
	}

	return ret, nil
}

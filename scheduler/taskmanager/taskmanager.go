package taskmanager

import (
	"mesos-sdk"
	"mesos-sdk/taskmngr"
	"stash.verizon.com/dkt/mlog"
)

type TaskState struct {
	taskinfo mesos.TaskInfo
	id       string
	status   string
}

// Satisfy the Task interface
func (t *TaskState) Info() (mesos.TaskInfo, error) {
	return t.taskinfo, nil
}

func (t *TaskState) SetInfo(taskinfo mesos.TaskInfo) error {
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
	tasks       map[string]taskmngr.Task
}

func NewManager() *Manager {
	return &Manager{tasks: make(map[string]taskmngr.Task)}
}

// Satisfy the TaskManager interface

// Provision a task
func (m *Manager) Add(task taskmngr.Task) error {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return err
	}

	m.tasks[id] = task

	return err
}

// Delete a task
func (m *Manager) Delete(id string) {
	delete(m.tasks, id)
}

// Set a task status
func (m *Manager) SetTask(task taskmngr.Task, status string) error {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return err
	}
	task.SetStatus(status)
	m.tasks[id] = task // Overwrite the old value here.
	return nil
}

// Check if a task has a particular status.
func (m *Manager) IsTask(task taskmngr.Task, status string) (bool, error) {
	id, err := task.Id()
	if err != nil {
		mlog.Error(err.Error())
		return false, err
	}
	if _, ok := m.tasks[id]; !ok {
		return false, nil
	}
	currentStatus, err := task.Status()
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

	if _, ok := m.tasks[id]; ok {
		return false, nil
	}
	return true, nil
}

// Check if we have tasks left to execute.
func (m *Manager) HasQueuedTasks() (bool, error) {
	mlog.Error("has any tasks?", len(m.tasks))
	return !(len(m.tasks) == 0), nil
}

func (m *Manager) Tasks() map[string]taskmngr.Task {
	return m.tasks
}

// Create a map of taskid->agentid
func TaskIdAgentIdMap(m map[string]taskmngr.Task) (map[string]string, error) {
	ret := make(map[string]string, len(m))
	for k, v := range m {
		taskInfo, err := v.Info()
		if err == nil {
			ret[k] = taskInfo.AgentID.Value
		}
	}

	return ret, nil
}

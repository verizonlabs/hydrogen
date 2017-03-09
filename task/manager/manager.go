package manager

import (
	"errors"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/structures"
)

/*
Satisfies the TaskManager interface in sdk
*/
type SprintTaskManager struct {
	tasks         *structures.ConcurrentMap
	launchedTasks map[string]*mesos_v1.TaskInfo
	queuedTasks   map[string]*mesos_v1.TaskInfo
}

func NewTaskManager(cmap *structures.ConcurrentMap) *SprintTaskManager {
	return &SprintTaskManager{
		tasks:         cmap,
		launchedTasks: make(map[string]*mesos_v1.TaskInfo),
		queuedTasks:   make(map[string]*mesos_v1.TaskInfo),
	}
}

func (m *SprintTaskManager) Add(task *mesos_v1.TaskInfo) {
	m.tasks.Set(task.GetName(), *task)
	m.SetTaskQueued(task)
}

func (m *SprintTaskManager) Delete(task *mesos_v1.TaskInfo) {
	m.tasks.Delete(task.GetName())
	delete(m.queuedTasks, task.GetName())
	delete(m.launchedTasks, task.GetName())
}

func (m *SprintTaskManager) Get(name *string) (*mesos_v1.TaskInfo, error) {
	ret := m.tasks.Get(*name)
	if ret != nil {
		r := ret.(mesos_v1.TaskInfo)
		return &r, nil
	}
	return &mesos_v1.TaskInfo{}, errors.New("Could not find task.")
}

func (m *SprintTaskManager) HasTask(task *mesos_v1.TaskInfo) bool {
	ret := m.tasks.Get(task.GetName())
	if ret == nil {
		return false
	}
	return true
}

func (m *SprintTaskManager) TotalTasks() int {
	return m.tasks.Length()
}

func (m *SprintTaskManager) Tasks() *structures.ConcurrentMap {
	return m.tasks
}

// Functions below are not part of interface.

func (m *SprintTaskManager) SetTaskQueued(task *mesos_v1.TaskInfo) error {
	if task != nil {
		m.queuedTasks[task.GetName()] = task
		return nil
	}
	return errors.New("Task passed into SetTaskQueued is nil")
}

func (m *SprintTaskManager) SetTaskLaunched(task *mesos_v1.TaskInfo) error {
	if _, ok := m.queuedTasks[task.GetName()]; ok {
		delete(m.queuedTasks, task.GetName())  // Delete it from queue.
		m.launchedTasks[task.GetName()] = task // Add to launched tasks.
		return nil
	}
	// This task isn't queued up, reject.
	return errors.New("Task is not in queue, cannot set to launching.")
}

func (m *SprintTaskManager) HasQueuedTasks() bool {
	if m.tasks.Length() == 0 {
		return false
	}
	for v := range m.tasks.Iterate() {
		task := v.Value.(mesos_v1.TaskInfo)
		if _, ok := m.queuedTasks[task.GetName()]; ok {
			return true
		}
	}
	return false
}

// Check to see if any tasks we have match the id passed in.
func (m *SprintTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	if m.tasks.Length() == 0 {
		return nil
	}
	for v := range m.tasks.Iterate() {
		task := v.Value.(mesos_v1.TaskInfo)
		if task.GetTaskId().GetValue() == id.GetValue() {
			return &task
		}
	}
	return nil
}

func (m *SprintTaskManager) QueuedTasks() map[string]*mesos_v1.TaskInfo {
	return m.queuedTasks
}

func (m *SprintTaskManager) LaunchedTasks() map[string]*mesos_v1.TaskInfo {
	return m.launchedTasks
}

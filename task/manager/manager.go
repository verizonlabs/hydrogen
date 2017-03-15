package manager

import (
	"errors"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task/manager"
)

/*
Satisfies the TaskManager interface in sdk
*/
type SprintTaskManager struct {
	tasks    *structures.ConcurrentMap
	launched map[string]*mesos_v1.TaskInfo
	queued   map[string]*mesos_v1.TaskInfo
}

func NewTaskManager(cmap *structures.ConcurrentMap) *SprintTaskManager {
	return &SprintTaskManager{
		tasks:    cmap,
		launched: make(map[string]*mesos_v1.TaskInfo),
		queued:   make(map[string]*mesos_v1.TaskInfo),
	}
}

func (m *SprintTaskManager) Add(task *mesos_v1.TaskInfo) {
	m.tasks.Set(task.GetName(), *task)
	m.setTaskQueued(task)
}

func (m *SprintTaskManager) Delete(task *mesos_v1.TaskInfo) {
	m.tasks.Delete(task.GetName())
	delete(m.queued, task.GetName())
	delete(m.launched, task.GetName())
}

func (m *SprintTaskManager) Get(name *string) (*mesos_v1.TaskInfo, error) {
	ret := m.tasks.Get(*name)
	if ret != nil {
		r := ret.(mesos_v1.TaskInfo)
		return &r, nil
	}
	return &mesos_v1.TaskInfo{}, errors.New("Could not find task.")
}

// Check to see if any tasks we have match the id passed in.
func (m *SprintTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	if m.tasks.Length() == 0 {
		return nil, errors.New("Task manager is empty.")
	}
	for v := range m.tasks.Iterate() {
		task := v.Value.(mesos_v1.TaskInfo)
		if task.GetTaskId().GetValue() == id.GetValue() {
			return &task, nil
		}
	}
	return nil, errors.New("Could not find task by id: " + id.GetValue())
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

func (m *SprintTaskManager) SetState(state *mesos_v1.TaskState, task *mesos_v1.TaskInfo) error {
	switch *state {
	case taskmanager.LAUNCHED:
		m.setTaskLaunched(task)
	case taskmanager.LOST:
		m.setTaskQueued(task)
	case taskmanager.DROPPED:
		m.setTaskQueued(task)
	case taskmanager.ERROR:
	case taskmanager.UNREACHABLE:
	case taskmanager.UNKNOWN:
	case taskmanager.STARTING:
	case taskmanager.STAGING:
	case taskmanager.FINISHED:
		m.Delete(task)
	case taskmanager.KILLED:
		m.Delete(task)
	case taskmanager.KILLING:
	case taskmanager.GONE:
		m.setTaskQueued(task)
	case taskmanager.GONE_BY_OPERATOR:
	case taskmanager.FAILED:
		m.setTaskQueued(task)
	}
	return nil
}

// Get's all tasks within a certain state.
func (m *SprintTaskManager) GetState(state *mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	// TODO: can refactor the for loops into a private method.
	ret := []*mesos_v1.TaskInfo{}
	switch *state {
	case taskmanager.LAUNCHED:
		for _, v := range m.launched {
			ret = append(ret, v)
		}
	case taskmanager.LOST:
	case taskmanager.DROPPED:
	case taskmanager.ERROR:
	case taskmanager.UNREACHABLE:
	case taskmanager.UNKNOWN:
	case taskmanager.STARTING:
	case taskmanager.STAGING:
		for _, v := range m.queued {
			ret = append(ret, v)
		}
	case taskmanager.FINISHED:
	case taskmanager.KILLED:
	case taskmanager.KILLING:
	case taskmanager.GONE:
	case taskmanager.GONE_BY_OPERATOR:
	case taskmanager.FAILED:
		for _, v := range m.queued {
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func (m *SprintTaskManager) setTaskQueued(task *mesos_v1.TaskInfo) error {
	if task != nil {
		delete(m.launched, task.GetName())
		m.queued[task.GetName()] = task
		return nil
	}
	return errors.New("Task passed into SetTaskQueued is nil")
}

func (m *SprintTaskManager) setTaskLaunched(task *mesos_v1.TaskInfo) error {
	if _, ok := m.queued[task.GetName()]; ok {
		delete(m.queued, task.GetName())  // Delete it from queue.
		m.launched[task.GetName()] = task // Add to launched tasks.
		return nil
	}
	// This task isn't queued up, reject.
	return errors.New("Task is not in queue, cannot set to launching.")
}

func (m *SprintTaskManager) hasQueuedTasks() bool {
	if m.tasks.Length() == 0 {
		return false
	}
	for v := range m.tasks.Iterate() {
		task := v.Value.(mesos_v1.TaskInfo)
		if _, ok := m.queued[task.GetName()]; ok {
			return true
		}
	}
	return false
}

func (m *SprintTaskManager) queuedTasks() map[string]*mesos_v1.TaskInfo {
	return m.queued
}

func (m *SprintTaskManager) launchedTasks() map[string]*mesos_v1.TaskInfo {
	return m.launched
}

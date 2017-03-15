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

type task struct {
	info  *mesos_v1.TaskInfo
	state mesos_v1.TaskState
}

type SprintTaskManager struct {
	tasks *structures.ConcurrentMap
}

func NewTaskManager(cmap *structures.ConcurrentMap) *SprintTaskManager {
	return &SprintTaskManager{
		tasks: cmap,
	}
}

func (m *SprintTaskManager) Add(t *mesos_v1.TaskInfo) {
	m.tasks.Set(t.GetName(), task{
		state: manager.UNKNOWN,
		info:  t,
	})
}

func (m *SprintTaskManager) Delete(task *mesos_v1.TaskInfo) {
	m.tasks.Delete(task.GetName())
}

func (m *SprintTaskManager) Get(name *string) (*mesos_v1.TaskInfo, error) {
	ret := m.tasks.Get(*name)
	if ret != nil {
		return ret.(task).info, nil
	}

	return nil, errors.New("Could not find task.")
}

// Check to see if any tasks we have match the id passed in.
func (m *SprintTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	if m.tasks.Length() == 0 {
		return nil, errors.New("Task manager is empty.")
	}

	for v := range m.tasks.Iterate() {
		task := v.Value.(task)
		if task.info.GetTaskId().GetValue() == id.GetValue() {
			return task.info, nil
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

func (m *SprintTaskManager) SetState(state mesos_v1.TaskState, t *mesos_v1.TaskInfo) {
	m.tasks.Set(t.GetName(), task{
		info:  t,
		state: state,
	})

	switch state {
	case manager.FINISHED:
		m.Delete(t)
	case manager.KILLED:
		m.Delete(t)
	}
}

// Get's all tasks within a certain state.
func (m *SprintTaskManager) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	tasks := []*mesos_v1.TaskInfo{}
	for v := range m.tasks.Iterate() {
		task := v.Value.(task)
		if task.state == state {
			tasks = append(tasks, task.info)
		}
	}

	if len(tasks) == 0 {
		return nil, errors.New("No tasks found with specified state of " + state.String())
	}

	return tasks, nil
}

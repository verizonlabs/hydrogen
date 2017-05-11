package manager

import (
	"encoding/json"
	"mesos-framework-sdk/include/mesos_v1"
	r "mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/task"
	t "mesos-framework-sdk/task/manager"
	"sprint/task/builder"
	"sprint/task/manager"
)

var (
	DEFAULT_RETRY_POLICY = &task.TimeRetry{
		Time:       "1.5",
		Backoff:    true,
		MaxRetries: 3,
	}
)

//api manager will hold refs to task/resource manager.

type (
	ApiManager interface {
		Deploy([]byte) (*mesos_v1.TaskInfo, error)
		Kill([]byte) error
		Update([]byte) (*mesos_v1.TaskInfo, error)
		Status(string) (mesos_v1.TaskState, error)
		AllTasks() ([]t.Task, error)
	}

	Manager struct {
		resourceManager r.ResourceManager
		taskManager     manager.SprintTaskManager
		scheduler       scheduler.Scheduler
	}
)

func NewApiManager(r r.ResourceManager, t manager.SprintTaskManager, s scheduler.Scheduler) *Manager {
	return &Manager{
		resourceManager: r,
		taskManager:     t,
		scheduler:       s,
	}
}

func (m *Manager) Deploy(decoded []byte) (*mesos_v1.TaskInfo, error) {
	var a task.ApplicationJSON
	err := json.Unmarshal(decoded, &a)

	mesosTask, err := builder.Application(&a)
	if err != nil {
		return nil, err
	}

	// If we have any filters, let the resource manager know.
	if len(a.Filters) > 0 {
		if err := m.resourceManager.AddFilter(mesosTask, a.Filters); err != nil {
			return nil, err
		}
	}

	if a.Retry != nil {
		err := m.taskManager.AddPolicy(a.Retry, mesosTask)
		if err != nil {
			return nil, err
		}
	} else {
		err := m.taskManager.AddPolicy(DEFAULT_RETRY_POLICY, mesosTask)
		if err != nil {
			return nil, err
		}
	}

	if err := m.taskManager.Add(mesosTask); err != nil {
		return nil, err
	}

	m.scheduler.Revive()
	return mesosTask, nil
}

func (m *Manager) Update(decoded []byte) (*mesos_v1.TaskInfo, error) {
	var a task.ApplicationJSON
	err := json.Unmarshal(decoded, &a)
	if err != nil {
		return nil, err
	}

	// Check if this task already exists
	taskToKill, err := m.taskManager.Get(&a.Name)
	if err != nil {
		return nil, err
	}

	mesosTask, err := builder.Application(&a)
	if err != nil {
		return nil, err
	}

	if a.Retry != nil {
		err = m.taskManager.AddPolicy(a.Retry, mesosTask)
	} else {
		err = m.taskManager.AddPolicy(DEFAULT_RETRY_POLICY, mesosTask)
	}

	if err != nil {
		return nil, err
	}

	err = m.taskManager.Set(t.UNKNOWN, mesosTask)
	if err != nil {
		return nil, err
	}

	m.scheduler.Kill(taskToKill.GetTaskId(), taskToKill.GetAgentId())
	m.scheduler.Revive()

	return mesosTask, nil
}

func (m *Manager) Kill(decoded []byte) error {
	var a task.KillJson
	err := json.Unmarshal(decoded, &a)
	if err != nil {
		return err
	}

	// Make sure we have a name to look up
	if a.Name == nil {
		return nil
	}

	// Look up task in task manager
	tsk, err := m.taskManager.Get(a.Name)
	if err != nil {
		return err
	}

	state, err := m.taskManager.State(tsk.Name)
	if err != nil {
		return err
	}

	err = m.taskManager.Delete(tsk)
	if err != nil {
		return err
	}

	m.resourceManager.ClearFilters(tsk)
	if *state == t.STAGING || *state == t.RUNNING || *state == t.STARTING {
		_, err := m.scheduler.Kill(tsk.GetTaskId(), tsk.GetAgentId())
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (m *Manager) Status(name string) (mesos_v1.TaskState, error) {
	state, err := m.taskManager.State(&name)
	if err != nil {
		return t.UNKNOWN, err
	}

	return *state, nil
}

func (m *Manager) AllTasks() ([]t.Task, error) {
	tasks, err := m.taskManager.All()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

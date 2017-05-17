package manager

import (
	"encoding/json"
	"errors"
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
	ApiParser interface {
		Deploy([]byte) (*mesos_v1.TaskInfo, error)
		Kill([]byte) error
		Update([]byte) (*mesos_v1.TaskInfo, error)
		Status(string) (mesos_v1.TaskState, error)
		AllTasks() ([]t.Task, error)
	}

	Parser struct {
		resourceManager r.ResourceManager
		taskManager     manager.SprintTaskManager
		scheduler       scheduler.Scheduler
	}
)

func NewApiParser(r r.ResourceManager, t manager.SprintTaskManager, s scheduler.Scheduler) *Parser {
	return &Parser{
		resourceManager: r,
		taskManager:     t,
		scheduler:       s,
	}
}

func (m *Parser) Deploy(decoded []byte) (*mesos_v1.TaskInfo, error) {
	var appJson task.ApplicationJSON
	err := json.Unmarshal(decoded, &appJson)

	mesosTask, err := builder.Application(&appJson)
	if err != nil {
		return nil, err
	}

	// Check if a task with this name already exists.
	exists, err := m.taskManager.Get(mesosTask.Name)
	if exists != nil {
		return nil, errors.New("Duplicate task name.")
	}

	// If we have any filters, let the resource manager know.
	if len(appJson.Filters) > 0 {
		if err := m.resourceManager.AddFilter(mesosTask, appJson.Filters); err != nil {
			return nil, err
		}
	}

	if appJson.Retry != nil {
		err := m.taskManager.AddPolicy(appJson.Retry, mesosTask)
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

func (m *Parser) Update(decoded []byte) (*mesos_v1.TaskInfo, error) {
	var appJson task.ApplicationJSON
	err := json.Unmarshal(decoded, &appJson)
	if err != nil {
		return nil, err
	}

	// Check if this task already exists
	taskToKill, err := m.taskManager.Get(&appJson.Name)
	if err != nil {
		return nil, err
	}

	mesosTask, err := builder.Application(&appJson)
	if err != nil {
		return nil, err
	}

	if appJson.Retry != nil {
		err = m.taskManager.AddPolicy(appJson.Retry, mesosTask)
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

func (m *Parser) Kill(decoded []byte) error {
	var appJson task.KillJson
	err := json.Unmarshal(decoded, &appJson)
	if err != nil {
		return err
	}

	// Make sure we have a name to look up
	if appJson.Name == nil {
		return nil
	}

	// Look up task in task manager
	tsk, err := m.taskManager.Get(appJson.Name)
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

func (m *Parser) Status(name string) (mesos_v1.TaskState, error) {
	state, err := m.taskManager.State(&name)
	if err != nil {
		return t.UNKNOWN, err
	}

	return *state, nil
}

func (m *Parser) AllTasks() ([]t.Task, error) {
	tasks, err := m.taskManager.All()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

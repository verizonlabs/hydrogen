package manager

import (
	"encoding/json"
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	r "mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/task"
	t "mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"sprint/task/builder"
	"sprint/task/control"
	"sprint/task/manager"
	"strconv"
)

type (
	ApiParser interface {
		Deploy([]byte) (*mesos_v1.TaskInfo, error)
		Kill([]byte) (string, error)
		Update([]byte) (*mesos_v1.TaskInfo, error)
		Status(string) (mesos_v1.TaskState, error)
		AllTasks() ([]t.Task, error)
	}

	Parser struct {
		ctrlPlane       control.ControlPlane
		resourceManager r.ResourceManager
		taskManager     manager.SprintTaskManager
		scheduler       scheduler.Scheduler
	}
)

// NewApiParser returns an object that marshalls JSON and handles the input from the API endpoints.
func NewApiParser(r r.ResourceManager, t manager.SprintTaskManager, s scheduler.Scheduler) *Parser {
	return &Parser{
		resourceManager: r,
		taskManager:     t,
		scheduler:       s,
	}
}

// Deploy takes a slice of bytes and marshals them into a Application json struct.
func (m *Parser) Deploy(decoded []byte) (*mesos_v1.TaskInfo, error) {
	var appJSON task.ApplicationJSON
	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return nil, err
	}

	mesosTask, err := builder.Application(&appJSON)
	if err != nil {
		return nil, err
	}

	//
	// If it's a group deployment, make a group object
	// group object can be passed around just like
	// a single task and updated on events.
	//
	// TaskLayer() <- updates and intents to launch are sent here
	// to check to see what type of task we have
	// a task can be:
	// one-off.
	// grouped
	// long-running (?)
	//
	//
	// NewTask(mesosTask) // This holds an abstract task
	// The New Task object then can be modified in the business layer.
	//
	// Updates to the task come back as a mesos task.
	// The task can be looked up and then updated accordingly.
	//
	//

	// Deployment strategy
	if appJSON.Instances == 1 {
		if err := m.taskManager.Add(mesosTask); err != nil {
			return nil, err
		}
		err := m.taskManager.AddPolicy(appJSON.Retry, mesosTask)
		if err != nil {
			return nil, err
		}
		if len(appJSON.Filters) > 0 {
			if err := m.resourceManager.AddFilter(mesosTask, appJSON.Filters); err != nil {
				return nil, err
			}
		}

	} else if appJSON.Instances > 1 {
		originalName := mesosTask.GetName()
		taskId := mesosTask.GetTaskId().GetValue()
		for i := 0; i < appJSON.Instances; i++ {
			duplicate := *mesosTask
			duplicate.Name = utils.ProtoString(originalName + "-" + strconv.Itoa(i+1))
			duplicate.TaskId = &mesos_v1.TaskID{Value: utils.ProtoString(taskId + "-" + strconv.Itoa(i+1))}
			if err := m.taskManager.Add(&duplicate); err != nil {
				return nil, err
			}
			err := m.taskManager.AddPolicy(appJSON.Retry, &duplicate)
			if err != nil {
				return nil, err
			}
			if len(appJSON.Filters) > 0 {
				if err := m.resourceManager.AddFilter(&duplicate, appJSON.Filters); err != nil {
					return nil, err
				}
			}
		}
	} else {
		return nil, errors.New("0 instances passed in. Not launching any tasks")
	}

	m.scheduler.Revive()
	return mesosTask, nil
}

// Update takes a slice of bytes and marshalls them into an ApplicationJSON struct.
func (m *Parser) Update(decoded []byte) (*mesos_v1.TaskInfo, error) {
	var appJSON task.ApplicationJSON
	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return nil, err
	}

	taskToKill, err := m.taskManager.Get(&appJSON.Name)
	if err != nil {
		return nil, err
	}

	mesosTask, err := builder.Application(&appJSON)
	if err != nil {
		return nil, err
	}

	if appJSON.Retry != nil {
		err = m.taskManager.AddPolicy(appJSON.Retry, mesosTask)
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

// Kill takes a slice of bytes and marshalls them into a kill json struct.
func (m *Parser) Kill(decoded []byte) (string, error) {
	var appJSON task.KillJson
	err := json.Unmarshal(decoded, &appJSON)
	if err != nil {
		return "", err
	}

	// Make sure we have a name to look up
	if appJSON.Name == nil {
		return "", errors.New("Task name is nil")
	}

	// Look up task in task manager
	tsk, err := m.taskManager.Get(appJSON.Name)
	if err != nil {
		return "", err
	}

	state, err := m.taskManager.State(tsk.Name)
	if err != nil {
		return "", err
	}

	err = m.taskManager.Delete(tsk)
	if err != nil {
		return "", err
	}

	m.resourceManager.ClearFilters(tsk)

	// If we are "unknown" that means the master doesn't know about the task, no need to make an HTTP call.
	if *state != t.UNKNOWN {
		_, err := m.scheduler.Kill(tsk.GetTaskId(), tsk.GetAgentId())
		if err != nil {
			return "", err
		}
	}

	return *appJSON.Name, nil
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

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
	"sprint/task/manager"
	"strconv"
)

//api manager will hold refs to task/resource manager.

type (
	ApiParser interface {
		Deploy([]byte) (*mesos_v1.TaskInfo, error)
		Kill([]byte) (string, error)
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

	// Deployment strategy
	if appJson.Instances == 1 {
		if err := m.taskManager.Add(mesosTask); err != nil {
			return nil, err
		}
		err := m.taskManager.AddPolicy(appJson.Retry, mesosTask)
		if err != nil {
			return nil, err
		}
		if len(appJson.Filters) > 0 {
			if err := m.resourceManager.AddFilter(mesosTask, appJson.Filters); err != nil {
				return nil, err
			}
		}

	} else if appJson.Instances > 1 {
		originalName := mesosTask.GetName()
		if err := m.taskManager.CreateGroup(originalName); err != nil {
			return nil, err
		} // Create our new group.
		if err := m.taskManager.SetSize(originalName, appJson.Instances); err != nil {
			return nil, err
		}
		taskId := mesosTask.GetTaskId().GetValue()
		for i := 0; i < appJson.Instances; i++ {
			duplicate := *mesosTask
			duplicate.Name = utils.ProtoString(originalName + "-" + strconv.Itoa(i+1))
			duplicate.TaskId = &mesos_v1.TaskID{Value: utils.ProtoString(taskId + "-" + strconv.Itoa(i+1))}
			if err := m.taskManager.Add(&duplicate); err != nil {
				return nil, err
			}
			err := m.taskManager.AddPolicy(appJson.Retry, &duplicate)
			if err != nil {
				return nil, err
			}
			if len(appJson.Filters) > 0 {
				if err := m.resourceManager.AddFilter(&duplicate, appJson.Filters); err != nil {
					return nil, err
				}
			}
		}
	} else {
		return nil, errors.New("0 instances passed in. Not launching any tasks.")
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

func (m *Parser) Kill(decoded []byte) (string, error) {
	var appJson task.KillJson
	err := json.Unmarshal(decoded, &appJson)
	if err != nil {
		return "", err
	}

	// Make sure we have a name to look up
	if appJson.Name == nil {
		return "", errors.New("Task name is nil")
	}

	// Look up task in task manager
	tsk, err := m.taskManager.Get(appJson.Name)
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
	} else {
		// we need to delete this task from a group since no event is generated
		// by mesos if we kill a task before it's launched. For example,
		// This occurs if we have a task that can't get a valid offer and sits in the queue.
		// but we still want to remove it from the queue anyways.
		if m.taskManager.IsInGroup(tsk) {
			m.taskManager.Unlink(tsk.GetName(), tsk.GetAgentId())
			m.taskManager.SetSize(tsk.GetName(), -1)
			m.taskManager.DeleteGroup(tsk.GetName())
		}
	}

	return *appJson.Name, nil
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

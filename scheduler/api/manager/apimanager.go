package apimanager

import (
	"encoding/json"
	"mesos-framework-sdk/include/mesos"
	r "mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/task"
	t "mesos-framework-sdk/task/manager"
	"sprint/task/builder"
	"sprint/task/manager"
)

var (
	DEFAULT_RETRY_POLICY = &task.TimeRetry{
		Time:    "1.5",
		Backoff: true,
	}
)

//api manager will hold refs to task/resource manager.

type (
	ApiManager interface {
		Deploy([]byte) (*mesos_v1.TaskInfo, error)
		Kill([]byte) error
		Update([]byte) (*mesos_v1.TaskInfo, error)
		Status(string) (mesos_v1.TaskState, error)
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
		m.taskManager.AddPolicy(a.Retry, mesosTask)
	}

	m.taskManager.Set(t.UNKNOWN, mesosTask)

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
	if a.Name != nil {
		// Look up task in task manager
		t, err := m.taskManager.Get(a.Name)
		if err != nil {
			return err
		}
		// Get all tasks in RUNNING state.
		running, _ := m.taskManager.GetState(mesos_v1.TaskState_TASK_RUNNING)
		// If we get an error, it means no tasks are currently in the running state.
		// We safely ignore this- the range over the empty list will be skipped regardless.

		// Check if our task is in the list of RUNNING tasks.
		for _, task := range running {
			// If it is, then send the kill signal.
			if task.GetName() == t.GetName() {
				// First Kill call to the mesos-master.
				_, err := m.scheduler.Kill(t.GetTaskId(), t.GetAgentId())
				if err != nil {
					// If it fails, try to kill it again.
					_, err := m.scheduler.Kill(t.GetTaskId(), t.GetAgentId())
					if err != nil {
						// We've tried twice and still failed.
						// Send back an error message.
						return err
					}
				}
				// Our kill call has worked, delete it from the task queue.
				m.taskManager.Delete(t)
				m.resourceManager.ClearFilters(t)
				// Response appropriately.
				return nil
			}
		}
		// If we get here, our task isn't in the list of RUNNING tasks.
		// Delete it from the queue regardless.
		// We run into this case if a task is flapping or unable to launch
		// or get an appropriate offer.
		m.taskManager.Delete(t)
		m.resourceManager.ClearFilters(t)
		return nil
	}
	return nil
}

func (m *Manager) Status(name string) (mesos_v1.TaskState, error) {
	_, err := m.taskManager.Get(&name)
	if err != nil {
		return t.UNKNOWN, err
	}
	queued, err := m.taskManager.GetState(t.STAGING)
	if err != nil {
		return t.UNKNOWN, err
	}

	for _, task := range queued {
		if task.GetName() == name {
			return t.STAGING, nil
		}
	}
	// TODO (tim): What about finished? Killed?
	return t.RUNNING, nil
}

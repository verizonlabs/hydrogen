package events

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	"time"
)

//
// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
//
func (s *SprintEventController) restoreTasks() {
	var tasks map[string]string
	var err error
	for {
		tasks, err = s.kv.ReadAll("/tasks")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get all task data: %s", err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		break
	}

	for _, value := range tasks {
		var task manager.Task
		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			s.logger.Emit(logging.ERROR, err.Error())
		}

		var b bytes.Buffer
		b.Write(data)
		d := gob.NewDecoder(&b)
		err = d.Decode(&task)
		if err != nil {
			s.logger.Emit(logging.ERROR, err.Error())
		}
		s.TaskManager().Set(task.State, task.Info)
	}
}

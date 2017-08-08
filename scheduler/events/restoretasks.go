package events

import (
	"encoding/json"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	sprint "sprint/task/manager"
	"strings"
)

//
// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
//
func (s *SprintEventController) restoreTasks() error {
	tasks, err := s.storage.ReadAll(sprint.TASK_DIRECTORY)
	if err != nil {
		return err
	}

	for _, value := range tasks {
		var task manager.Task
		if err != nil {
			s.logger.Emit(logging.ERROR, err.Error())
		}

		dec := json.NewDecoder(strings.NewReader(value))
		if err = dec.Decode(&task); err != nil {
			return err
		}
		if err := s.TaskManager().Add(&task); err != nil {
			return err
		}
	}

	return nil
}

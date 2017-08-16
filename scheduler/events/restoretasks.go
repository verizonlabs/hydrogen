// Copyright 2017 Verizon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"mesos-framework-sdk/task/manager"
	sprint "sprint/task/manager"
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
		task, err := new(manager.Task).Decode([]byte(value))
		if err != nil {
			return err
		}

		if err := s.TaskManager().Add(task); err != nil {
			return err
		}
	}

	return nil
}

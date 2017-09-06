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
	"mesos-framework-sdk/include/mesos_v1"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/utils"
)

func (d *ExecutorController) Launch(launch *exec.Event_Launch) {
	d.logger.Emit(logging.INFO, "Launch event received: ", *launch.Task)
	err := d.executor.Update(&mesos_v1.TaskStatus{
		TaskId: launch.Task.TaskId,
		State:  mesos_v1.TaskState_TASK_RUNNING.Enum(),
		Uuid:   utils.Uuid(),
		Source: mesos_v1.TaskStatus_SOURCE_EXECUTOR.Enum(),
	})
	if err != nil {
		d.logger.Emit(logging.ERROR, "ERROR DURING UPDATE OF TASK", err.Error())
	}
}

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
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Update(t *testing.T) {
	ctrl := workingEventController()

	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
				State:  state,
			}})
	}
}

func TestSprintEventController_UpdateWithNilTaskId(t *testing.T) {
	ctrl := workingEventController()
	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: nil},
				State:  state,
			}})
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: nil,
				State:  state,
			}})
	}
	ctrl.Update(nil)
}

func TestSprintEventController_UpdateWithInvalidState(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Update(&mesos_v1_scheduler.Event_Update{
		Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: nil},
			State:  nil,
		}})
	ctrl.Update(&mesos_v1_scheduler.Event_Update{
		Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: nil},
			State:  mesos_v1.TaskState(100).Enum(), // Doesn't exist.
		}})
}

func TestSprintEventController_UpdateWith(t *testing.T) {
	ctrl := brokenSchedulerEventController()
	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
				State:  state,
			}})
	}
}

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
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
)

//
// Failure is a public method that respond to a Failure event sent by the mesos master.
// We log the failure here with the logger.
//
func (s *SprintEventController) Failure(fail *mesos_v1_scheduler.Event_Failure) {
	if fail != nil {
		s.logger.Emit(logging.ERROR, "Executor %s failed with status %d", fail.GetExecutorId().GetValue(), fail.GetStatus())
	} else {
		s.logger.Emit(logging.ERROR, "Recieved an nil failure message!")
	}
}

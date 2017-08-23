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
// Error is a public method that handles an error event sent by the mesos master.
// We log errors to the logger, and handle any special cases as needed here.
//
func (s *SprintEventController) Error(err *mesos_v1_scheduler.Event_Error) {
	s.logger.Emit(logging.ERROR, "Error event recieved: %v", err)
}

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
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
)

func (d *SprintExecutorController) Acknowledged(acknowledge *exec.Event_Acknowledged) {
	// The executor is expected to maintain a list of status updates not acknowledged by the agent via the ACKNOWLEDGE events.
	// The executor is expected to maintain a list of tasks that have not been acknowledged by the agent.
	// A task is considered acknowledged if at least one of the status updates for this task is acknowledged by the agent.
	d.logger.Emit(logging.INFO, "Acknowledge event received")
}

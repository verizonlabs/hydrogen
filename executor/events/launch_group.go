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
	exec "github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_executor"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
)

func (d *Handler) LaunchGroup(launchGroup *exec.Event_LaunchGroup) {
	d.logger.Emit(logging.INFO, "Launch_group event received")
}

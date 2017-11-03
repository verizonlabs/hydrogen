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
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
)

//
// Rescind is a public method that handles a rescind event from the mesos-master.
// Rescind events only occur if an offer isn't declined properly within the offer
// timeout period.
//
func (e *Router) Rescind(rescindEvent *mesos_v1_scheduler.Event_Rescind) {
	if rescindEvent != nil {
		e.logger.Emit(logging.INFO, "Rescind event recieved: %v", *rescindEvent)
	} else {
		e.logger.Emit(logging.INFO, "Rescind event recieved was nil!")
	}
}

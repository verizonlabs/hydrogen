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
// Message is a public method that handles a message event from the mesos master.
// We simply print this message to the log, however in the future we may utilize
// the message passing functionality to send custom logging errors.
//
func (e *Event) Message(msg *mesos_v1_scheduler.Event_Message) {
	if msg != nil {
		e.controller.Logger.Emit(logging.INFO, "Message event recieved: %v", *msg)
	} else {
		e.controller.Logger.Emit(logging.ERROR, "Recieved a nil message!")
	}
}

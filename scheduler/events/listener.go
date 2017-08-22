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

// Main event loop that listens on channels forever until framework terminates.
// The Listen() function handles events coming in on the events channel.
// This acts as a type of switch board to route events to their proper callback methods.
func (e *Event) Listen(c chan *mesos_v1_scheduler.Event) {
	for {
		select {
		case t := <-c:
			switch t.GetType() {
			case mesos_v1_scheduler.Event_SUBSCRIBED:
				e.Subscribed(t.GetSubscribed())
			case mesos_v1_scheduler.Event_ERROR:
				e.Error(t.GetError())
			case mesos_v1_scheduler.Event_FAILURE:
				e.Failure(t.GetFailure())
			case mesos_v1_scheduler.Event_INVERSE_OFFERS:
				e.InverseOffer(t.GetInverseOffers())
			case mesos_v1_scheduler.Event_MESSAGE:
				e.Message(t.GetMessage())
			case mesos_v1_scheduler.Event_OFFERS:
				e.Offers(t.GetOffers())
			case mesos_v1_scheduler.Event_RESCIND:
				e.Rescind(t.GetRescind())
			case mesos_v1_scheduler.Event_RESCIND_INVERSE_OFFER:
				e.RescindInverseOffer(t.GetRescindInverseOffer())
			case mesos_v1_scheduler.Event_UPDATE:
				e.Update(t.GetUpdate())
			case mesos_v1_scheduler.Event_HEARTBEAT:
				e.controller.RefreshFrameworkIdLease()
			case mesos_v1_scheduler.Event_UNKNOWN:
				e.controller.Logger.Emit(logging.ALARM, "Unknown event received")
			}
		case r := <-e.controller.Revive:
			e.reschedule(r)
		}
	}
}

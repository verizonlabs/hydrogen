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
	scheduler "mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
)

//
// Main event loop that listens on channels forever until framework terminates.
// The Listen() function handles events coming in on the events channel.
// This acts as a type of switch board to route events to their proper callback methods.
//
func (s *SprintEventController) Listen() {
	for {
		select {
		case t := <-s.events:
			switch t.GetType() {
			case scheduler.Event_SUBSCRIBED:
				s.Subscribed(t.GetSubscribed())
			case scheduler.Event_ERROR:
				s.Error(t.GetError())
			case scheduler.Event_FAILURE:
				s.Failure(t.GetFailure())
			case scheduler.Event_INVERSE_OFFERS:
				s.InverseOffer(t.GetInverseOffers())
			case scheduler.Event_MESSAGE:
				s.Message(t.GetMessage())
			case scheduler.Event_OFFERS:
				s.Offers(t.GetOffers())
			case scheduler.Event_RESCIND:
				s.Rescind(t.GetRescind())
			case scheduler.Event_RESCIND_INVERSE_OFFER:
				s.RescindInverseOffer(t.GetRescindInverseOffer())
			case scheduler.Event_UPDATE:
				s.Update(t.GetUpdate())
			case scheduler.Event_HEARTBEAT:
				s.refreshFrameworkIdLease()
			case scheduler.Event_UNKNOWN:
				s.logger.Emit(logging.ALARM, "Unknown event received")
			}
		case revive := <-s.revive:
			s.reschedule(revive)
		}
	}
}

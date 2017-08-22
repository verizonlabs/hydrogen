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
	mockResourceManager "mesos-framework-sdk/resources/manager/test"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestEvent_Listen(t *testing.T) {
	ch := make(chan *mesos_v1_scheduler.Event)
	ctrl := workingEventController()
	go ctrl.Run(ch)

	ch <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}

	e := NewEvent(workingEventController(), new(mockResourceManager.MockResourceManager))
	go e.Listen(ch)
	// Test all event messages.
	ch <- &mesos_v1_scheduler.Event{
		Type:  mesos_v1_scheduler.Event_ERROR.Enum(),
		Error: &mesos_v1_scheduler.Event_Error{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_FAILURE.Enum(),
		Failure: &mesos_v1_scheduler.Event_Failure{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:          mesos_v1_scheduler.Event_INVERSE_OFFERS.Enum(),
		InverseOffers: &mesos_v1_scheduler.Event_InverseOffers{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_MESSAGE.Enum(),
		Message: &mesos_v1_scheduler.Event_Message{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:   mesos_v1_scheduler.Event_OFFERS.Enum(),
		Offers: &mesos_v1_scheduler.Event_Offers{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:    mesos_v1_scheduler.Event_RESCIND.Enum(),
		Rescind: &mesos_v1_scheduler.Event_Rescind{},
	}
	ch <- &mesos_v1_scheduler.Event{
		Type:   mesos_v1_scheduler.Event_UPDATE.Enum(),
		Update: &mesos_v1_scheduler.Event_Update{},
	}
}

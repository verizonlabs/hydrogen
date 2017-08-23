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

func TestSprintEventController_InverseOffer(t *testing.T) {
	ctrl := workingEventController()
	ctrl.InverseOffer(&mesos_v1_scheduler.Event_InverseOffers{
		InverseOffers: []*mesos_v1.InverseOffer{
			{
				Id:             &mesos_v1.OfferID{Value: utils.ProtoString("id")},
				FrameworkId:    &mesos_v1.FrameworkID{Value: utils.ProtoString("id")},
				Unavailability: &mesos_v1.Unavailability{Start: &mesos_v1.TimeInfo{Nanoseconds: utils.ProtoInt64(0)}},
			},
		},
	})
}

func TestSprintEventController_InverseOfferWithNilOffer(t *testing.T) {
	ctrl := workingEventController()
	ctrl.InverseOffer(&mesos_v1_scheduler.Event_InverseOffers{InverseOffers: nil})
	ctrl.InverseOffer(nil)
}

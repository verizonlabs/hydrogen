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
	"mesos-framework-sdk/utils"
)

//
// Decline offers is a private method to organize a list of offers that are to be declined by the
// scheduler.
//
func (s *SprintEventController) declineOffers(offers []*mesos_v1.Offer, refuseSeconds float64) {
	if len(offers) == 0 {
		return
	}

	declineIDs := make([]*mesos_v1.OfferID, 0, len(offers))

	// Decline whatever offers are left over
	for _, id := range offers {
		declineIDs = append(declineIDs, id.GetId())
	}

	s.scheduler.Decline(declineIDs, &mesos_v1.Filters{RefuseSeconds: utils.ProtoFloat64(refuseSeconds)})
}

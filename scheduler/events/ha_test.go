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

import "testing"

// Possible to test communication with a variadic input for a channel.
// If a channel is passed in, it communicates on that channel.
// Can block on the "read" of that channel.
func TestSprintEventController_Communicate(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.Communicate() // this never gets covered since this test closes too fast
	ctrl.Election()
}

func TestSprintEventController_Election(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Election()
}

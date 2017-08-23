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

// Can we create a leader?
func TestSprintEventController_CreateLeader(t *testing.T) {
	ctrl := workingEventController()
	if err := ctrl.CreateLeader(); err != nil {
		t.Logf("Failed to create a leader %v", err.Error())
		t.Fail()
	}
}

// Can we create and get the leader?
func TestSprintEventController_GetLeader(t *testing.T) {
	ctrl := workingEventController()
	if err := ctrl.CreateLeader(); err != nil {
		t.Logf("Failed to create a leader %v\n", err.Error())
		t.Fail()
	}
	// Leader string is simply an empty string during testing...
	_, err := ctrl.GetLeader()
	if err != nil {
		t.Logf("Failed to grab the leader %v\n", err.Error())
		t.Fail()
	}
}

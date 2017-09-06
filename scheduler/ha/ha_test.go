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

package ha

import (
	mockLogger "mesos-framework-sdk/logging/test"
	"hydrogen/scheduler"
	mockStorage "hydrogen/task/persistence/test"
	"testing"
)

// Possible to test communication with a variadic input for a channel.
// If a channel is passed in, it communicates on that channel.
// Can block on the "read" of that channel.
func TestHA_Communicate(t *testing.T) {
	ha := NewHA(new(mockStorage.MockStorage), new(mockLogger.MockLogger), &scheduler.LeaderConfiguration{
		IP: "1", // Make sure we break out of our HA loop by matching on what mock storage gives us.
	})
	go ha.Communicate()
	ha.Election()
}

func TestHA_Election(t *testing.T) {
	ha := NewHA(new(mockStorage.MockStorage), new(mockLogger.MockLogger), &scheduler.LeaderConfiguration{
		IP: "1", // Make sure we break out of our HA loop by matching on what mock storage gives us.
	})
	ha.Election()
}

// Can we create a leader?
func TestHA_CreateLeader(t *testing.T) {
	ha := NewHA(new(mockStorage.MockStorage), new(mockLogger.MockLogger), &scheduler.LeaderConfiguration{
		IP: "1", // Make sure we break out of our HA loop by matching on what mock storage gives us.
	})
	if err := ha.CreateLeader(); err != nil {
		t.Logf("Failed to create a leader %v", err.Error())
		t.Fail()
	}
}

// Can we create and get the leader?
func TestHA_GetLeader(t *testing.T) {
	ha := NewHA(new(mockStorage.MockStorage), new(mockLogger.MockLogger), &scheduler.LeaderConfiguration{
		IP: "1", // Make sure we break out of our HA loop by matching on what mock storage gives us.
	})
	if err := ha.CreateLeader(); err != nil {
		t.Logf("Failed to create a leader %v\n", err.Error())
		t.Fail()
	}
	// Leader string is simply an empty string during testing...
	_, err := ha.GetLeader()
	if err != nil {
		t.Logf("Failed to grab the leader %v\n", err.Error())
		t.Fail()
	}
}

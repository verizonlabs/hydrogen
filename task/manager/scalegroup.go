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

package manager

import (
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
)

type (
	ScaleGrouping interface {
		CreateGroup(string) error               // Creates a new task group
		DeleteGroup(string) error               // Deletes an entire task group
		SetSize(string, int) error              // Can handle positive/negative values.
		Link(string, *mesos_v1.AgentID) error   // Links an agent from a task group.
		Unlink(string, *mesos_v1.AgentID) error // Unlinks an agent from a task group
		ReadGroup(string) []*mesos_v1.AgentID   // Gathers all agents tied to a task group
		IsInGroup(*mesos_v1.TaskInfo) bool      // Checks for existence of a taskgroup
	}
)

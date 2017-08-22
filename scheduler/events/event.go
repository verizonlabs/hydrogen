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
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler/events"
	"sprint/scheduler/controller"
)

// Event contains various event handlers and holds data that callbacks need to access/modify.
type Event struct {
	controller      *controller.EventController
	resourceManager manager.ResourceManager
}

// NewEvent returns a new Event type which adheres to the SchedulerEvent interface.
func NewEvent(c *controller.EventController, r manager.ResourceManager) events.SchedulerEvent {
	return &Event{
		controller:      c,
		resourceManager: r,
	}
}

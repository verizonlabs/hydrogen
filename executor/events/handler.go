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
	"github.com/verizonlabs/mesos-framework-sdk/executor"
	"github.com/verizonlabs/mesos-framework-sdk/executor/events"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_executor"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
)

type Handler struct {
	executor executor.Executor
	logger   logging.Logger
}

// NewEvent returns a new Event type which adheres to the SchedulerEvent interface.
func NewHandler(e executor.Executor, l logging.Logger) events.ExecutorEvents {
	return &Handler{
		executor: e,
		logger:   l,
	}
}

// Run executes the appropriate callback based on the event type.
func (h *Handler) Run(event *mesos_v1_executor.Event) {
	switch event.GetType() {
	case mesos_v1_executor.Event_SUBSCRIBED:
		h.Subscribed(event.GetSubscribed())
	case mesos_v1_executor.Event_ACKNOWLEDGED:
		h.Acknowledged(event.GetAcknowledged())
	case mesos_v1_executor.Event_MESSAGE:
		h.Message(event.GetMessage())
	case mesos_v1_executor.Event_KILL:
		h.Kill(event.GetKill())
	case mesos_v1_executor.Event_LAUNCH:
		h.Launch(event.GetLaunch())
	case mesos_v1_executor.Event_LAUNCH_GROUP:
		h.LaunchGroup(event.GetLaunchGroup())
	case mesos_v1_executor.Event_SHUTDOWN:
		h.Shutdown()
	case mesos_v1_executor.Event_ERROR:
		h.Error(event.GetError())
	case mesos_v1_executor.Event_UNKNOWN:
		h.logger.Emit(logging.ALARM, "Unknown event received")
	}
}

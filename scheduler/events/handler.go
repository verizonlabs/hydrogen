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
	sched "github.com/verizonlabs/hydrogen/scheduler"
	"github.com/verizonlabs/hydrogen/task/persistence"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
	resourceManager "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler/events"
	taskManager "github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/plan"
	"os"
	"sync"
)

// Router updates and routes messages from mesos.
type Router struct {
	planner plan.PlanQueue
	logger  logging.Logger
}

// NewEvent returns a new Event type which adheres to the SchedulerEvent interface.
func NewEventRouter(planner plan.PlanQueue, l logging.Logger) events.SchedulerEvent {
	return &Router{
		logger:  l,
		planner: planner,
	}
}

// Run executes the appropriate callback based on the event type.
func (r *Router) Run(event *mesos_v1_scheduler.Event) {
	// Update the current plan.

	switch event.GetType() {
	case mesos_v1_scheduler.Event_SUBSCRIBED:
		h.Subscribed(event.GetSubscribed())
	case mesos_v1_scheduler.Event_ERROR:
		h.Error(event.GetError())
	case mesos_v1_scheduler.Event_FAILURE:
		h.Failure(event.GetFailure())
	case mesos_v1_scheduler.Event_INVERSE_OFFERS:
		h.InverseOffer(event.GetInverseOffers())
	case mesos_v1_scheduler.Event_MESSAGE:
		h.Message(event.GetMessage())
	case mesos_v1_scheduler.Event_OFFERS:
		h.Offers(event.GetOffers())
	case mesos_v1_scheduler.Event_RESCIND:
		h.Rescind(event.GetRescind())
	case mesos_v1_scheduler.Event_RESCIND_INVERSE_OFFER:
		h.RescindInverseOffer(event.GetRescindInverseOffer())
	case mesos_v1_scheduler.Event_UPDATE:
		h.Update(event.GetUpdate())
	case mesos_v1_scheduler.Event_HEARTBEAT:
		h.refreshFrameworkIdLease()
	case mesos_v1_scheduler.Event_UNKNOWN:
		h.logger.Emit(logging.ALARM, "Unknown event received")
	}
}

func (h *Router) Signals() {
	h.RLock()
	if h.frameworkLease == 0 {
		os.Exit(0)
	}
	h.RUnlock()

	// Refresh our lease before we die so that we start an accurate countdown.
	err := h.refreshFrameworkIdLease()
	if err != nil {
		h.logger.Emit(logging.ERROR, "Failed to refresh leader lease before exiting: %s", err.Error())
		os.Exit(6)
	}

	os.Exit(0)
}

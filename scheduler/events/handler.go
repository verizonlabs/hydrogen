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
	sched "hydrogen/scheduler"
	"hydrogen/task/persistence"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	resourceManager "mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/scheduler/events"
	taskManager "mesos-framework-sdk/task/manager"
	"os"
	"sync"
)

// Handler contains various event handlers and holds data that callbacks need to access/modify.
type Handler struct {
	taskManager     taskManager.TaskManager
	resourceManager resourceManager.ResourceManager
	config          *sched.Configuration
	scheduler       scheduler.Scheduler
	storage         persistence.Storage
	revive          chan *taskManager.Task
	logger          logging.Logger
	frameworkLease  int64
	sync.RWMutex
}

// NewEvent returns a new Event type which adheres to the SchedulerEvent interface.
func NewHandler(t taskManager.TaskManager,
	r resourceManager.ResourceManager,
	c *sched.Configuration,
	s scheduler.Scheduler,
	o persistence.Storage,
	v chan *taskManager.Task,
	l logging.Logger) events.SchedulerEvent {

	return &Handler{
		taskManager:     t,
		resourceManager: r,
		config:          c,
		scheduler:       s,
		storage:         o,
		revive:          v,
		logger:          l,
	}
}

// Run executes the appropriate callback based on the event type.
func (h *Handler) Run(event *mesos_v1_scheduler.Event) {
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

func (h *Handler) Signals() {
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

// Refreshes the lifetime of our persisted framework ID.
func (h *Handler) refreshFrameworkIdLease() error {
	policy := h.storage.CheckPolicy(nil)
	return h.storage.RunPolicy(policy, func() error {
		h.RLock()
		err := h.storage.RefreshLease(h.frameworkLease)
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to refresh framework ID lease: %s", err.Error())
		}
		h.RUnlock()

		return err
	})
}

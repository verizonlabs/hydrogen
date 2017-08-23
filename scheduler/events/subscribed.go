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
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	"os"
)

// Subscribed is a public method that handles subscription events from the master.
// We handle our subscription by writing the framework ID to storage.
// Then we gather all of our launched non-terminal tasks and reconcile them explicitly.
func (h *Handler) Subscribed(subEvent *mesos_v1_scheduler.Event_Subscribed) {
	id := subEvent.GetFrameworkId()
	idVal := id.GetValue()
	h.scheduler.FrameworkInfo().Id = id
	h.logger.Emit(logging.INFO, "Subscribed with an ID of %s", idVal)

	err := h.createFrameworkIdLease(idVal)
	if err != nil {
		h.logger.Emit(logging.ERROR, "Failed to persist leader information: %s", err.Error())
		os.Exit(3)
	}

	h.scheduler.Revive() // Reset to revive offers regardless if there are tasks or not.
	// We do this to force a check for any tasks that we might have missed during downtime.
	// Reconcile after we subscribe in case we resubscribed due to a failure.

	// Get all launched non-terminal tasks.
	launched, err := h.taskManager.AllByState(manager.RUNNING)
	if err != nil {
		h.logger.Emit(logging.INFO, "Not reconciling: %s", err.Error())
		return
	}

	toReconcile := make([]*mesos_v1.TaskInfo, 0, len(launched))
	for _, t := range launched {
		toReconcile = append(toReconcile, t.Info)
	}

	h.scheduler.Reconcile(toReconcile)
}

// Create and persist our framework ID with an attached lifetime.
func (h *Handler) createFrameworkIdLease(idVal string) error {
	policy := h.storage.CheckPolicy(nil)
	return h.storage.RunPolicy(policy, func() error {
		lease, err := h.storage.CreateWithLease("/frameworkId", idVal, int64(h.scheduler.FrameworkInfo().GetFailoverTimeout()))
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
			return err
		}

		h.frameworkLease = lease

		return nil
	})
}

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

package controller

import (
	"github.com/verizonlabs/hydrogen/scheduler/ha"
	"github.com/verizonlabs/hydrogen/task/manager"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler/events"
	sdkTaskManager "github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/plan"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	frameworkIDKey = "/frameworkId"
)

// The event controller is responsible for coordinating:
// High Availability (HA) between schedulers in a cluster.
// Subscribing to the mesos master.
// Talking to the etcd cluster.
// Listening for event from the mesos master
// Responding to events.
//
// The event controller acts as a main switchboard for our events.
// Scheduler calls are made by the scheduler class.

type (

	// EventController is the primary controller that coordinates the calls from the scheduler with the events received from Mesos.
	// The controller is what is "run" to kick off our scheduler and subscribe to Mesos.
	// Additionally, our framework's high availability is coordinated by this controller.
	EventController struct {
		planner plan.PlanQueue
		logger  logging.Logger
		ha      *ha.HA
	}
)

// Returns the main controller that's used to coordinate the calls/events from/to the scheduler.
func NewEventController(planner plan.PlanQueue, logger logging.Logger, ha *ha.HA) *EventController {

	return &EventController{
		planner: planner,
		logger:  logger,
		ha:      ha,
	}
}

// TODO (tim): This can be refactored into a "Subscribe" plan.
// Subscribe is a special plan since it blocks forever during the listening stage.
//

//
// Main Run() function serves to run all the necessary logic
// to set up the event controller to subscribe, and listen to events from
// the mesos master in the cluster.
// This method blocks forever, or until the scheduler is brought down.
//
func (s *EventController) Run() {
	var currentPlan plan.Plan

	// LOOP START
	for {
		// See if we have a plan to run
		currentPlan = s.planner.Peek()

		// If not, just block and wait for a new plan to be put on the stack.
		if currentPlan.Type() == plan.Idle {
			// block until we have something to run.
			currentPlan = <-s.planner.Wait()
		}

		// Check if the plan is already running
		if currentPlan.Status() != plan.RUNNING {
			// If not, run the plan.
			// Run the current plan once we have one to run that's not idle.
			// The plan gets updated in Execute automatically
			currentPlan.Execute()
		} else {

		}

		// Check status and make sure it returns successful, if not,
		// treat the failure as necessary by updating the plan.
		if currentPlan.Status() == plan.FINISHED {
			s.planner.Pop()
		}
	}
	// LOOP END

	// TODO (tim): Refactor this into the subscribe plan.
	//s.listen(events, revives, handler)
}

// Listens for Mesos events, tasks that need to be revived, and signals and routes to the appropriate handler.
func (s *EventController) listen(c chan *mesos_v1_scheduler.Event, r chan *sdkTaskManager.Task, h events.SchedulerEvent) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case event := <-c:
			h.Run(event)
		case task := <-r:
			h.Reschedule(task)
		case <-sigs:
			h.Signals()
		}
	}
}

//
// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
//
func (s *EventController) restoreTasks() error {
	tasks, err := s.storage.ReadAll(manager.TASK_DIRECTORY)
	if err != nil {
		return err
	}

	for _, value := range tasks {
		task, err := new(sdkTaskManager.Task).Decode([]byte(value))
		if err != nil {
			return err
		}

		s.taskManager.Restore(task)
	}

	return nil
}

// Keep our state in check by periodically reconciling.
func (s *EventController) periodicReconcile() {
	ticker := time.NewTicker(s.config.Scheduler.ReconcileInterval)

	for {
		select {
		case <-ticker.C:
			recon, err := s.taskManager.AllByState(sdkTaskManager.RUNNING)
			if err != nil {
				continue
			}
			toReconcile := []*mesos_v1.TaskInfo{}
			for _, t := range recon {
				toReconcile = append(toReconcile, t.Info)
			}
			_, err = s.scheduler.Reconcile(toReconcile)
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to reconcile all running tasks: %s", err.Error())
			}
		}
	}
}

// Set our framework ID in memory from what's currently persisted.
// The framework ID is needed by the scheduler for almost any call to Mesos.
func (s *EventController) setFrameworkId() error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		id, err := s.storage.Read(frameworkIDKey)
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set the framework ID: %s", err.Error())
			return err
		}

		s.scheduler.FrameworkInfo().Id = &mesos_v1.FrameworkID{Value: &id}
		return nil
	})
}

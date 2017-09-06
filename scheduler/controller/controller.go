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
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	sdkScheduler "mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/scheduler/events"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"os"
	"os/signal"
	sprintSched "hydrogen/scheduler"
	"hydrogen/scheduler/ha"
	sprintTaskManager "hydrogen/task/manager"
	"hydrogen/task/persistence"
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
		config      *sprintSched.Configuration
		scheduler   sdkScheduler.Scheduler
		taskManager sdkTaskManager.TaskManager
		storage     persistence.Storage
		logger      logging.Logger
		ha          *ha.HA
	}
)

// Returns the main controller that's used to coordinate the calls/events from/to the scheduler.
func NewEventController(
	config *sprintSched.Configuration,
	scheduler sdkScheduler.Scheduler,
	manager sdkTaskManager.TaskManager,
	storage persistence.Storage,
	logger logging.Logger,
	ha *ha.HA) *EventController {

	return &EventController{
		config:      config,
		scheduler:   scheduler,
		taskManager: manager,
		storage:     storage,
		logger:      logger,
		ha:          ha,
	}
}

//
// Main Run() function serves to run all the necessary logic
// to set up the event controller to subscribe, and listen to events from
// the mesos master in the cluster.
// This method blocks forever, or until the scheduler is brought down.
//
func (s *EventController) Run(events chan *mesos_v1_scheduler.Event, revives chan *sdkTaskManager.Task, handler events.SchedulerEvent) {

	// Start the election.
	s.logger.Emit(logging.INFO, "Starting leader election socket server")
	go s.ha.Communicate()

	// Block here until we either become a leader or a standby.
	// If we are the leader we break out and continue to execute the rest of the scheduler.
	// If we are a standby then we connect to the leader and wait for the process to start over again.
	s.ha.Election()

	// Get the frameworkId from etcd and set it to our frameworkID in our struct.
	err := s.setFrameworkId()
	if err != nil {
		s.logger.Emit(logging.ERROR, "Failed to get the framework ID from persistent storage: %s", err.Error())
	}

	// Recover our state (if any) in the event we (or the server) go down.
	s.logger.Emit(logging.INFO, "Restoring any persisted state from data store")
	err = s.restoreTasks()
	if err != nil {
		s.logger.Emit(logging.INFO, "Failed to restore persisted state: %s", err.Error())
		os.Exit(2)
	}

	// Kick off our scheduled reconciling.
	s.logger.Emit(logging.INFO, "Starting periodic reconciler thread with a %g minute interval", s.config.Scheduler.ReconcileInterval.Minutes())
	go s.periodicReconcile()

	go func() {
		for {
			leader, err := s.ha.GetLeader()
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to get leader information: %s", err.Error())
				os.Exit(5)
			}

			// We should only ever reach here if we hit a network partition and the standbys lose connection to the leader.
			// If this happens we need to check if there really is another leader alive that we just can't reach.
			// If we wrongly think we are the leader and try to subscribe when there's already a leader then we will disconnect the leader.
			// Both the leader and the incorrectly determined new leader will continue to disconnect each other.
			if leader != s.config.Leader.IP {
				s.logger.Emit(logging.ERROR, "We are not the leader so we should not be subscribing. "+
					"This is most likely caused by a network partition between the leader and standbys")
				os.Exit(1)
			}

			resp, err := s.scheduler.Subscribe(events)
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to subscribe: %s", err.Error())
				if resp != nil && resp.StatusCode == 401 {
					// TODO define constants and clean up all areas that currently exit with a hardcoded value.
					os.Exit(8)
				}

				time.Sleep(time.Duration(s.config.Scheduler.SubscribeRetry))
			}
		}
	}()

	s.listen(events, revives, handler)
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
	tasks, err := s.storage.ReadAll(sprintTaskManager.TASK_DIRECTORY)
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

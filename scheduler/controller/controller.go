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
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"os"
	"os/signal"
	sprintSched "sprint/scheduler"
	"sprint/scheduler/ha"
	sprintTaskManager "sprint/task/manager"
	"sprint/task/persistence"
	"sync"
	"syscall"
	"time"
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
		Config         *sprintSched.Configuration
		Scheduler      sdkScheduler.Scheduler
		TaskManager    sdkTaskManager.TaskManager
		Storage        persistence.Storage
		Logger         logging.Logger
		Ha             *ha.HA
		FrameworkLease int64
		lock           sync.RWMutex
		Revive         chan *sdkTaskManager.Task
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
		Config:      config,
		Scheduler:   scheduler,
		TaskManager: manager,
		Storage:     storage,
		Logger:      logger,
		Ha:          ha,
		Revive:      make(chan *sdkTaskManager.Task),
	}
}

//
// Main Run() function serves to run all the necessary logic
// to set up the event controller to subscribe, and listen to events from
// the mesos master in the cluster.
// This method blocks forever, or until the scheduler is brought down.
//
func (s *EventController) Run(events chan *mesos_v1_scheduler.Event) {
	s.registerShutdownHandlers()

	// Start the election.
	s.Logger.Emit(logging.INFO, "Starting leader election socket server")
	go s.Ha.Communicate()

	// Block here until we either become a leader or a standby.
	// If we are the leader we break out and continue to execute the rest of the scheduler.
	// If we are a standby then we connect to the leader and wait for the process to start over again.
	s.Ha.Election()

	// Get the frameworkId from etcd and set it to our frameworkID in our struct.
	err := s.setFrameworkId()
	if err != nil {
		s.Logger.Emit(logging.ERROR, "Failed to get the framework ID from persistent storage: %s", err.Error())
	}

	// Recover our state (if any) in the event we (or the server) go down.
	s.Logger.Emit(logging.INFO, "Restoring any persisted state from data store")
	err = s.restoreTasks()
	if err != nil {
		s.Logger.Emit(logging.INFO, "Failed to restore persisted state: %s", err.Error())
		os.Exit(2)
	}

	// Kick off our scheduled reconciling.
	s.Logger.Emit(logging.INFO, "Starting periodic reconciler thread with a %g minute interval", s.Config.Scheduler.ReconcileInterval.Minutes())
	go s.periodicReconcile()

	go func() {
		for {
			leader, err := s.Ha.GetLeader()
			if err != nil {
				s.Logger.Emit(logging.ERROR, "Failed to get leader information: %s", err.Error())
				os.Exit(5)
			}

			// We should only ever reach here if we hit a network partition and the standbys lose connection to the leader.
			// If this happens we need to check if there really is another leader alive that we just can't reach.
			// If we wrongly think we are the leader and try to subscribe when there's already a leader then we will disconnect the leader.
			// Both the leader and the incorrectly determined new leader will continue to disconnect each other.
			if leader != s.Config.Leader.IP {
				s.Logger.Emit(logging.ERROR, "We are not the leader so we should not be subscribing. "+
					"This is most likely caused by a network partition between the leader and standbys")
				os.Exit(1)
			}

			resp, err := s.Scheduler.Subscribe(events)
			if err != nil {
				s.Logger.Emit(logging.ERROR, "Failed to subscribe: %s", err.Error())
				if resp != nil && resp.StatusCode == 401 {
					// TODO define constants and clean up all areas that currently exit with a hardcoded value.
					os.Exit(8)
				}

				time.Sleep(time.Duration(s.Config.Scheduler.SubscribeRetry))
			}
		}
	}()
}

//
// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
//
func (s *EventController) restoreTasks() error {
	tasks, err := s.Storage.ReadAll(sprintTaskManager.TASK_DIRECTORY)
	if err != nil {
		return err
	}

	for _, value := range tasks {
		task, err := new(sdkTaskManager.Task).Decode([]byte(value))
		if err != nil {
			return err
		}

		if err := s.TaskManager.Add(task); err != nil {
			return err
		}
	}

	return nil
}

// Keep our state in check by periodically reconciling.
func (s *EventController) periodicReconcile() {
	ticker := time.NewTicker(s.Config.Scheduler.ReconcileInterval)

	for {
		select {
		case <-ticker.C:
			recon, err := s.TaskManager.AllByState(sdkTaskManager.RUNNING)
			if err != nil {
				continue
			}
			toReconcile := []*mesos_v1.TaskInfo{}
			for _, t := range recon {
				toReconcile = append(toReconcile, t.Info)
			}
			_, err = s.Scheduler.Reconcile(toReconcile)
			if err != nil {
				s.Logger.Emit(logging.ERROR, "Failed to reconcile all running tasks: %s", err.Error())
			}
		}
	}
}

// Handle appropriate signals for graceful shutdowns.
func (s *EventController) registerShutdownHandlers() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs

		// Refresh our lease before we die so that we start an accurate countdown.
		s.lock.RLock()
		defer s.lock.RUnlock()
		if s.FrameworkLease != 0 {
			err := s.RefreshFrameworkIdLease()
			if err != nil {
				s.Logger.Emit(logging.ERROR, "Failed to refresh leader lease before exiting: %s", err.Error())
				os.Exit(6)
			}
		}

		os.Exit(0)
	}()
}

// Set our framework ID in memory from what's currently persisted.
// The framework ID is needed by the scheduler for almost any call to Mesos.
func (s *EventController) setFrameworkId() error {
	policy := s.Storage.CheckPolicy(nil)
	return s.Storage.RunPolicy(policy, func() error {
		id, err := s.Storage.Read("/frameworkId")
		if err != nil {
			s.Logger.Emit(logging.ERROR, "Failed to set the framework ID: %s", err.Error())
			return err
		}

		s.Scheduler.FrameworkInfo().Id = &mesos_v1.FrameworkID{Value: &id}
		return nil
	})
}

// Create and persist our framework ID with an attached lifetime.
func (s *EventController) CreateFrameworkIdLease(idVal string) error {
	policy := s.Storage.CheckPolicy(nil)
	return s.Storage.RunPolicy(policy, func() error {
		lease, err := s.Storage.CreateWithLease("/frameworkId", idVal, int64(s.Scheduler.FrameworkInfo().GetFailoverTimeout()))
		if err != nil {
			s.Logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
			return err
		}

		s.lock.Lock()
		s.FrameworkLease = lease
		s.lock.Unlock()

		return nil
	})
}

// Refreshes the lifetime of our persisted framework ID.
func (s *EventController) RefreshFrameworkIdLease() error {
	policy := s.Storage.CheckPolicy(nil)
	return s.Storage.RunPolicy(policy, func() error {
		s.lock.RLock()
		err := s.Storage.RefreshLease(s.FrameworkLease)
		if err != nil {
			s.Logger.Emit(logging.ERROR, "Failed to refresh framework ID lease: %s", err.Error())
		}

		s.lock.RUnlock()

		return err
	})
}

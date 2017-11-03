package plan

import (
	scheduler2 "github.com/verizonlabs/hydrogen/scheduler"
	"github.com/verizonlabs/hydrogen/scheduler/ha"
	"github.com/verizonlabs/hydrogen/task/persistence"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	manager2 "hydrogen/task/manager"
	"mesos-framework-sdk/scheduler"
	"os"
	"sync"
	"time"
)

const (
	frameworkIDKey = "/frameworkId"
)

type SubscribePlan struct {
	Plan
	state           PlanState
	tasks           []*manager.Task
	taskManager     manager.TaskManager
	resourceManager resource.ResourceManager
	scheduler       scheduler.Scheduler
	storage         persistence.Storage
	ha              ha.HighAvailablity
	logger          logging.Logger
	config          scheduler2.Configuration
	events          chan *mesos_v1_scheduler.Event
	frameworkLease  int64
	sync.Mutex
}

func NewSubscribePlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	scheduler scheduler.Scheduler,
	storage persistence.Storage,
	ha ha.HighAvailablity,
	logger logging.Logger,
	config scheduler2.Configuration,
	events chan *mesos_v1_scheduler.Event) Plan {

	return SubscribePlan{
		tasks:           tasks,
		taskManager:     taskManager,
		resourceManager: resourceManager,
		scheduler:       scheduler,
		ha:              ha,
		logger:          logger,
		config:          config,
		events:          events,
	}
}

func (s *SubscribePlan) Execute() error {
	s.state = RUNNING
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
	s.logger.Emit(logging.INFO, "Starting periodic reconciler thread with a %g minute interval",
		s.config.Scheduler.ReconcileInterval.Minutes())

	// Reconcile is done by pushing the reconcile plan onto the queue instead.
	//go s.periodicReconcile()

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

			resp, err := s.scheduler.Subscribe(s.events)
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

	// Once we're done with setting up the HA election, we want to listen for updates.
	go s.listen(s.events)

	return nil
} // Execute the current plan.

func (s *SubscribePlan) Update(state PlanState) {
	s.state = state
} // Update the current plan state.

func (s *SubscribePlan) Status() PlanState {
	return s.state
} // Current status of the plan.

func (s *SubscribePlan) Type() PlanType {
	return Subscribe
} // Tells you the plan type.

// Set our framework ID in memory from what's currently persisted.
// The framework ID is needed by the scheduler for almost any call to Mesos.
func (s *SubscribePlan) setFrameworkId() error {
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

//
// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
//
func (s *SubscribePlan) restoreTasks() error {
	tasks, err := s.storage.ReadAll(manager2.TASK_DIRECTORY)
	if err != nil {
		return err
	}

	for _, value := range tasks {
		task, err := new(manager.Task).Decode([]byte(value))
		if err != nil {
			return err
		}

		s.taskManager.Restore(task)
	}

	return nil
}

// Listens for Mesos events, tasks that need to be revived, and signals and routes to the appropriate handler.
func (s *SubscribePlan) listen(c chan *mesos_v1_scheduler.Event) {
	// Handle signals outside of any plans in the plan queue.
	//sigs := make(chan os.Signal)
	//signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until we hear an update on this plan, or timeout.
	event := <-c

	switch event.GetType() {
	case mesos_v1_scheduler.Event_SUBSCRIBED:
		s.subscribedEvent(event.GetSubscribed())
	case mesos_v1_scheduler.Event_ERROR:
	case mesos_v1_scheduler.Event_FAILURE:
	case mesos_v1_scheduler.Event_INVERSE_OFFERS:
	case mesos_v1_scheduler.Event_MESSAGE:
	case mesos_v1_scheduler.Event_OFFERS:
	case mesos_v1_scheduler.Event_RESCIND:
	case mesos_v1_scheduler.Event_RESCIND_INVERSE_OFFER:
	case mesos_v1_scheduler.Event_UPDATE:
	case mesos_v1_scheduler.Event_HEARTBEAT:
	case mesos_v1_scheduler.Event_UNKNOWN:
		s.logger.Emit(logging.ALARM, "Unknown event received")
	}

}

func (s *SubscribePlan) subscribedEvent(subEvent *mesos_v1_scheduler.Event_Subscribed) {
	id := subEvent.GetFrameworkId()
	idVal := id.GetValue()
	s.scheduler.FrameworkInfo().Id = id
	s.logger.Emit(logging.INFO, "Subscribed with an ID of %s", idVal)

	err := s.createFrameworkIdLease(idVal)
	if err != nil {
		s.logger.Emit(logging.ERROR, "Failed to persist leader information: %s", err.Error())
		os.Exit(3)
	}

	s.scheduler.Revive() // Reset to revive offers regardless if there are tasks or not.
	// We do this to force a check for any tasks that we might have missed during downtime.
	// Reconcile after we subscribe in case we resubscribed due to a failure.

	// Get all launched non-terminal tasks.
	launched, err := s.taskManager.AllByState(manager.RUNNING)
	if err != nil {
		s.logger.Emit(logging.INFO, "Not reconciling: %s", err.Error())
		return
	}

	toReconcile := make([]*mesos_v1.TaskInfo, 0, len(launched))
	for _, t := range launched {
		toReconcile = append(toReconcile, t.Info)
	}

	s.scheduler.Reconcile(toReconcile)
}

// Refreshes the lifetime of our persisted framework ID.
func (s *SubscribePlan) refreshFrameworkIdLease() error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		s.Lock()
		err := s.storage.RefreshLease(s.frameworkLease)
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to refresh framework ID lease: %s", err.Error())
		}
		s.Unlock()

		return err
	})
}

// Create and persist our framework ID with an attached lifetime.
func (s *SubscribePlan) createFrameworkIdLease(idVal string) error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		lease, err := s.storage.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.FrameworkInfo().GetFailoverTimeout()))
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
			return err
		}

		s.Lock()
		s.frameworkLease = lease
		s.Unlock()

		return nil
	})
}

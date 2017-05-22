package events

import (
	"mesos-framework-sdk/ha"
	sched "mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/scheduler/events"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"os"

	"os/signal"
	sprintSched "sprint/scheduler"
	sprintTask "sprint/task/manager"
	"sprint/task/persistence"
	"sync"
	"syscall"
	"time"
)

//
// The event controller is responsible for handling:
// High Availability (HA) between schedulers in a cluster.
// Subscribing to the mesos master.
// Talking to the etcd cluster.
// Listening for event from the mesos master
// Responding to events.
//
// The event controller acts as a main switchboard for our events.
// Scheduler calls are made by the scheduler class.
//

const (
	// Note (tim): Is there a reasonable non-linear equation to determine refuse seconds?
	// f^2/num_of_nodes_in_cluster where f is # of tasks to handle at once (per offer cycle).
	refuseSeconds = 30.0 // Setting this to 30 as a "reasonable default".
)

type (
	// Provides pluggable controllers that the scheduler can interface with.
	// Also used extensively for testing with mocks.
	EventController interface {
		events.SchedulerEvent
		ha.Node
	}

	// Primary controller that coordinates the calls from the scheduler with the events received from Mesos.
	// The controller is what is "run" to kick off our scheduler and subscribe to Mesos.
	// Additionally, our framework's high availability is coordinated by this controller.
	SprintEventController struct {
		config          *sprintSched.Configuration
		scheduler       scheduler.Scheduler
		taskmanager     sprintTask.SprintTaskManager
		resourcemanager manager.ResourceManager
		events          chan *sched.Event
		storage         persistence.Storage
		logger          logging.Logger
		frameworkLease  int64
		status          ha.Status
		name            string
		lock            sync.RWMutex
	}
)

// Returns the main controller that's used to coordinate the calls/events from/to the scheduler.
func NewSprintEventController(
	config *sprintSched.Configuration,
	scheduler scheduler.Scheduler,
	manager sprintTask.SprintTaskManager,
	resourceManager manager.ResourceManager,
	eventChan chan *sched.Event,
	storage persistence.Storage,
	logger logging.Logger) EventController {

	return &SprintEventController{
		config:          config,
		taskmanager:     manager,
		scheduler:       scheduler,
		events:          eventChan,
		resourcemanager: resourceManager,
		storage:         storage,
		logger:          logger,
		status:          ha.Election,
		name:            utils.UuidAsString(),
	}
}

// Returns the scheduler that makes calls to Mesos.
func (s *SprintEventController) Scheduler() scheduler.Scheduler {
	return s.scheduler
}

// Returns the task manager that coordinates state.
func (s *SprintEventController) TaskManager() sprintTask.SprintTaskManager {
	return s.taskmanager
}

// Returns the resource manager that handles matching offers with tasks.
func (s *SprintEventController) ResourceManager() manager.ResourceManager {
	return s.resourcemanager
}

//
// Main Run() function serves to run all the necessary logic
// to set up the event controller to subscribe, and listen to events from
// the mesos master in the cluster.
// This method blocks forever, or until the scheduler is brought down.
//
func (s *SprintEventController) Run() {
	s.registerShutdownHandlers()

	// Start the election.
	s.logger.Emit(logging.INFO, "Starting leader election socket server")
	s.status = ha.Listening
	go s.Communicate()

	// Block here until we either become a leader or a standby.
	// If we are the leader we break out and continue to execute the rest of the scheduler.
	// If we are a standby then we connect to the leader and wait for the process to start over again.
	s.Election()

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
			leader, err := s.GetLeader()
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

			_, err = s.scheduler.Subscribe(s.events)
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to subscribe: %s", err.Error())
				time.Sleep(time.Duration(s.config.Scheduler.SubscribeRetry))
			}
		}
	}()

	select {
	case e := <-s.events:
		s.Subscribed(e.GetSubscribed())
	}
	s.Listen()
}

// Keep our state in check by periodically reconciling.
func (s *SprintEventController) periodicReconcile() {
	ticker := time.NewTicker(s.config.Scheduler.ReconcileInterval)
	for {
		select {
		case <-ticker.C:
			recon, err := s.TaskManager().AllByState(sdkTaskManager.RUNNING)
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to reconcile all running tasks: %s", err.Error())
				continue
			}
			s.Scheduler().Reconcile(recon)
		}
	}
}

// Handle appropriate signals for graceful shutdowns.
func (s *SprintEventController) registerShutdownHandlers() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs

		// Refresh our lease before we die so that we start an accurate countdown.
		s.lock.RLock()
		defer s.lock.RUnlock()
		if s.frameworkLease != 0 {
			err := s.refreshFrameworkIdLease()
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to refresh leader lease before exiting: %s", err.Error())
				os.Exit(6)
			}
		}

		os.Exit(0)
	}()
}

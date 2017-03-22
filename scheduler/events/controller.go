package events

/*
Adapted from mesos-framework-sdk
*/
import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources"
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"os"
	sprintSched "sprint/scheduler"
	sprintTaskManager "sprint/task/manager"
	"time"
)

const subscribeRetry = 2

type SprintEventController struct {
	config          *sprintSched.SchedulerConfiguration
	scheduler       *scheduler.DefaultScheduler
	taskmanager     *sprintTaskManager.SprintTaskManager
	resourcemanager *manager.DefaultResourceManager
	events          chan *sched.Event
	kv              *etcd.Etcd
	logger          logging.Logger
}

// TODO tons of params, make this cleaner...?
func NewSprintEventController(config *sprintSched.SchedulerConfiguration, scheduler *scheduler.DefaultScheduler, manager *sprintTaskManager.SprintTaskManager, resourceManager *manager.DefaultResourceManager, eventChan chan *sched.Event, kv *etcd.Etcd, logger logging.Logger) *SprintEventController {
	return &SprintEventController{
		config:          config,
		taskmanager:     manager,
		scheduler:       scheduler,
		events:          eventChan,
		resourcemanager: resourceManager,
		kv:              kv,
		logger:          logger,
	}
}

// Getter functions
func (s *SprintEventController) Scheduler() *scheduler.DefaultScheduler {
	return s.scheduler
}

// Getter function
func (s *SprintEventController) TaskManager() *sprintTaskManager.SprintTaskManager {
	return s.taskmanager
}

// Getter function
func (s *SprintEventController) ResourceManager() *manager.DefaultResourceManager {
	return s.resourcemanager
}

// Atomically set leader information.
func (s *SprintEventController) CreateLeader() {
	if err := s.kv.Create("/leader", s.config.LeaderIP); err != nil {
		s.logger.Emit(logging.ERROR, "Failed to set leader information: "+err.Error())
	}
}

// Atomically get leader information.
func (s *SprintEventController) GetLeader() (string, error) {
	leader, err := s.kv.Read("/leader")
	if err != nil {
		return "", err
	}

	return leader, nil
}

// TODO think about renaming this to subscribed since the scheduler from the SDK is really handling the subscribe call.
func (s *SprintEventController) Subscribe(subEvent *sched.Event_Subscribed) {
	id := subEvent.GetFrameworkId()
	idVal := id.GetValue()
	s.scheduler.Info.Id = id
	s.logger.Emit(logging.INFO, "Subscribed with an ID of %s", idVal)

	// TODO this is only called once assuming we subscribe and don't fail soon afterwards
	// TODO it's possible that we can fail long after our lease length
	// TODO if this is the case our failover in Mesos will start long after the lease has expired
	// TODO we should continually refresh this lease on every heart beat event
	if err := s.kv.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.Info.GetFailoverTimeout()), false); err != nil {
		s.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
	}

	// Get all launched non-terminal tasks.
	launched, err := s.taskmanager.GetState(sdkTaskManager.RUNNING)
	if err != nil {
		s.logger.Emit(logging.INFO, "Not reconciling: %s", err.Error())
		return
	}

	// Reconcile after we subscribe in case we resubscribed due to a failure.
	s.scheduler.Reconcile(launched)
}

func (s *SprintEventController) Run() {
	id, err := s.kv.Read("/frameworkId")
	if err == nil {
		s.scheduler.Info.Id = &mesos_v1.FrameworkID{Value: &id}
	}

	go func() {
		for {
			leader, err := s.kv.Read("/leader")
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to find the leader: %s", err.Error())
				os.Exit(1)
			}

			// We should only ever reach here if we hit a network partition and the standbys lose connection to the leader.
			// If this happens we need to check if there really is another leader alive that we just can't reach.
			// If we wrongly think we are the leader and try to subscribe when there's already a leader then we will disconnect the leader.
			// Both the leader and the incorrectly determined new leader will continue to disconnect each other.
			if leader != s.config.LeaderIP {
				s.logger.Emit(logging.ERROR, "We are not the leader so we should not be subscribing")
				s.logger.Emit(logging.ERROR, "This is most likely caused by a network partition between the leader and standbys")
				os.Exit(1)
			}

			err = s.scheduler.Subscribe(s.events)
			if err != nil {
				s.logger.Emit(logging.ERROR, "Failed to subscribe: %s", err.Error())
				time.Sleep(time.Duration(subscribeRetry) * time.Second)
			}
		}
	}()

	select {
	case e := <-s.events:
		s.Subscribe(e.GetSubscribed())
	}
	s.Listen()
}

// Main event loop that listens on channels forever until framework terminates.
func (s *SprintEventController) Listen() {
	for {
		select {
		case t := <-s.events:
			switch t.GetType() {
			case sched.Event_SUBSCRIBED:
				go s.Subscribe(t.GetSubscribed())
			case sched.Event_ERROR:
				go s.Error(t.GetError())
			case sched.Event_FAILURE:
				go s.Failure(t.GetFailure())
			case sched.Event_INVERSE_OFFERS:
				go s.InverseOffer(t.GetInverseOffers())
			case sched.Event_MESSAGE:
				go s.Message(t.GetMessage())
			case sched.Event_OFFERS:
				go s.Offers(t.GetOffers())
			case sched.Event_RESCIND:
				go s.Rescind(t.GetRescind())
			case sched.Event_RESCIND_INVERSE_OFFER:
				go s.RescindInverseOffer(t.GetRescindInverseOffer())
			case sched.Event_UPDATE:
				go s.Update(t.GetUpdate())
			case sched.Event_HEARTBEAT:
			case sched.Event_UNKNOWN:
				s.logger.Emit(logging.ALARM, "Unknown event received")
			}
		}
	}
}

func (s *SprintEventController) Offers(offerEvent *sched.Event_Offers) {
	queued, err := s.taskmanager.GetState(sdkTaskManager.UNKNOWN)
	if err != nil {
		s.logger.Emit(logging.INFO, "No tasks to launch.")

		var ids []*mesos_v1.OfferID
		for _, v := range offerEvent.GetOffers() {
			ids = append(ids, v.GetId())
		}

		s.scheduler.Decline(ids, nil)
		s.scheduler.Suppress()

		return
	}

	// Update our resources in the manager
	s.resourcemanager.AddOffers(offerEvent.GetOffers())

	offerIDs := []*mesos_v1.OfferID{}
	operations := []*mesos_v1.Offer_Operation{}

	for _, mesosTask := range queued {

		// See if we have resources.
		if s.resourcemanager.HasResources() {
			offer, err := s.resourcemanager.Assign(mesosTask)
			if err != nil {
				// It didn't match any offers.
				s.logger.Emit(logging.ERROR, err.Error())
				continue // We should decline.
			}

			t := &mesos_v1.TaskInfo{
				Name:      mesosTask.Name,
				TaskId:    mesosTask.GetTaskId(),
				AgentId:   offer.GetAgentId(),
				Command:   mesosTask.GetCommand(),
				Container: mesosTask.GetContainer(),
				Resources: mesosTask.GetResources(),
			}

			// TODO investigate this state further as it might cause side effects.
			// TODO this is artifically set to STAGING, it does not correspond to when Mesos sets this task as STAGING.
			// TODO for example other parts of the codebase may check for STAGING and this would cause it to be set too early.
			s.TaskManager().Set(sdkTaskManager.STAGING, t)

			offerIDs = append(offerIDs, offer.Id)
			operations = append(operations, resources.LaunchOfferOperation([]*mesos_v1.TaskInfo{t}))
		}
	}
	s.scheduler.Accept(offerIDs, operations, nil)
}

func (s *SprintEventController) Rescind(rescindEvent *sched.Event_Rescind) {
	s.logger.Emit(logging.INFO, "Rescind event recieved: %v", *rescindEvent)
	rescindEvent.GetOfferId().GetValue()
}

func (s *SprintEventController) Update(updateEvent *sched.Event_Update) {
	s.logger.Emit(logging.INFO, updateEvent.GetStatus().GetMessage())
	task, err := s.taskmanager.GetById(updateEvent.GetStatus().GetTaskId())
	if err != nil {
		// This task doesn't exist according to task manager so we can't update it's status.
		s.logger.Emit(logging.ERROR, err.Error())
		return
	}

	state := updateEvent.GetStatus().GetState()
	s.taskmanager.Set(state, task)

	// We only know our real state once we get an update about our task.
	// This also keeps our state in memory in sync with our persistent data store.
	// TODO this should be broken out somewhere, maybe in the task manager once it handles persistence.
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	err = e.Encode(sdkTaskManager.Task{
		Info:  task,
		State: state,
	})

	if err != nil {
		s.logger.Emit(logging.ERROR, "Failed to encode task")
	}

	id := task.TaskId.GetValue()
	// TODO: We should refactor the storage stuff to utilize the storage interface.
	if err := s.kv.Create("/tasks/"+id, base64.StdEncoding.EncodeToString(b.Bytes())); err != nil {
		s.logger.Emit(logging.ERROR, "Failed to save task %s with name %s to persistent data store", id, task.GetName())
	}

	// TODO we can probably remove a bunch of these cases since most of them only set the status like we do above.
	switch state {
	case mesos_v1.TaskState_TASK_FAILED:
		// TODO: Check task manager for task retry policy, then retry as given.
	case mesos_v1.TaskState_TASK_STAGING:
		// NOP, keep task set to "launched".
	case mesos_v1.TaskState_TASK_DROPPED:
		// Transient error, we should retry launching. Taskinfo is fine.
	case mesos_v1.TaskState_TASK_ERROR:
		// TODO: Error with the taskinfo sent to the agent. Give verbose reasoning back.
	case mesos_v1.TaskState_TASK_FINISHED:
		s.taskmanager.Delete(task)
		id := task.TaskId.GetValue()
		if err := s.kv.Delete("/tasks/" + id); err != nil {
			s.logger.Emit(logging.INFO, "Failed to delete task %s with name %s from persistent data store", id, task.GetName())
		}
	case mesos_v1.TaskState_TASK_GONE:
		// Agent is dead and task is lost.
	case mesos_v1.TaskState_TASK_GONE_BY_OPERATOR:
		// Agent might be dead, master is unsure. Will return to RUNNING state possibly or die.
	case mesos_v1.TaskState_TASK_KILLED:
		// Task was killed.
		s.taskmanager.Delete(task)
		id := task.TaskId.GetValue()
		if err := s.kv.Delete("/task/" + id); err != nil {
			s.logger.Emit(logging.INFO, "Failed to delete task %s with name %s from persistent data store", id, task.GetName())
		}
	case mesos_v1.TaskState_TASK_KILLING:
		// Task is in the process of catching a SIGNAL and shutting down.
	case mesos_v1.TaskState_TASK_LOST:
		// Task is unknown to the master and lost. Should reschedule.
	case mesos_v1.TaskState_TASK_RUNNING:
	case mesos_v1.TaskState_TASK_STARTING:
		// Task is still starting up. NOOP
	case mesos_v1.TaskState_TASK_UNKNOWN:
		// Task is unknown to the master. Should ignore.
	case mesos_v1.TaskState_TASK_UNREACHABLE:
		// Agent lost contact with master, could be a network error. No guarantee the task is still running.
		// Should we reschedule after waiting a certain peroid of time?
	default:
	}

	status := updateEvent.GetStatus()
	s.scheduler.Acknowledge(status.GetAgentId(), status.GetTaskId(), status.GetUuid())
}

func (s *SprintEventController) Message(msg *sched.Event_Message) {
	s.logger.Emit(logging.INFO, "Message event recieved: %v", *msg)
}

func (s *SprintEventController) Failure(fail *sched.Event_Failure) {
	s.logger.Emit(logging.ERROR, "Executor %s failed with status %d", fail.GetExecutorId().GetValue(), fail.GetStatus())
}

func (s *SprintEventController) Error(err *sched.Event_Error) {
	s.logger.Emit(logging.INFO, "Error event recieved: %v", err)
}

func (s *SprintEventController) InverseOffer(ioffers *sched.Event_InverseOffers) {
	s.logger.Emit(logging.INFO, "Inverse Offer event recieved: %v", ioffers)
}

func (s *SprintEventController) RescindInverseOffer(rioffers *sched.Event_RescindInverseOffer) {
	s.logger.Emit(logging.INFO, "Rescind Inverse Offer event recieved: %v", rioffers)
}

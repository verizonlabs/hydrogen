package eventcontroller

/*
Adapted from mesos-framework-sdk
*/
import (
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources"
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/task/manager"
	"time"
)

const subscribeRetry = 2

type SprintEventController struct {
	scheduler       *scheduler.DefaultScheduler
	taskmanager     *task_manager.DefaultTaskManager
	resourcemanager *manager.DefaultResourceManager
	events          chan *sched.Event
	kv              *etcd.Etcd
	logger          logging.Logger
}

func NewSprintEventController(scheduler *scheduler.DefaultScheduler, manager *task_manager.DefaultTaskManager, resourceManager *manager.DefaultResourceManager, eventChan chan *sched.Event, kv *etcd.Etcd, logger logging.Logger) *SprintEventController {
	return &SprintEventController{
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
func (s *SprintEventController) TaskManager() *task_manager.DefaultTaskManager {
	return s.taskmanager
}

// Getter function
func (s *SprintEventController) ResourceManager() *manager.DefaultResourceManager {
	return s.resourcemanager
}

func (s *SprintEventController) Subscribe(subEvent *sched.Event_Subscribed) {
	id := subEvent.GetFrameworkId()
	idVal := id.GetValue()
	s.scheduler.Info.Id = id
	s.logger.Emit(logging.INFO, "Subscribed with an ID of %s", idVal)

	if err := s.kv.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.Info.GetFailoverTimeout())); err != nil {
		s.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
	}
}

func (s *SprintEventController) Run() {
	id, err := s.kv.Read("/frameworkId")
	if err == nil {
		s.scheduler.Info.Id = &mesos_v1.FrameworkID{Value: &id}
	}

	go func() {
		for {
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

	// Check task manager for any active tasks.
	if s.taskmanager.HasQueuedTasks() {
		// Update our resources in the manager
		s.resourcemanager.AddOffers(offerEvent.GetOffers())

		offerIDs := []*mesos_v1.OfferID{}
		operations := []*mesos_v1.Offer_Operation{}

		for _, mesosTask := range s.taskmanager.QueuedTasks() {
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

				s.TaskManager().SetTaskLaunched(t)

				offerIDs = append(offerIDs, offer.Id)
				operations = append(operations, resources.LaunchOfferOperation([]*mesos_v1.TaskInfo{t}))

				data := proto.MarshalTextString(t)
				id := t.TaskId.GetValue()
				if err := s.kv.Create("/task/"+id, data); err != nil {
					s.logger.Emit(logging.ERROR, "Failed to save task %s with name %s to persistent data store", id, t.GetName())
				}
			}
		}
		s.scheduler.Accept(offerIDs, operations, nil)
	} else {
		var ids []*mesos_v1.OfferID
		for _, v := range offerEvent.GetOffers() {
			ids = append(ids, v.GetId())
		}
		// Decline and supress offers until we're ready again.
		s.logger.Emit(logging.INFO, "Declining %d offers", len(ids))
		s.scheduler.Decline(ids, nil) // We want to make sure all offers are declined.
		s.scheduler.Suppress()
	}
}

func (s *SprintEventController) Rescind(rescindEvent *sched.Event_Rescind) {
	s.logger.Emit(logging.INFO, "Rescind event recieved.: %v", *rescindEvent)
	rescindEvent.GetOfferId().GetValue()
}

func (s *SprintEventController) Update(updateEvent *sched.Event_Update) {
	task := s.taskmanager.GetById(updateEvent.GetStatus().GetTaskId())
	if task != nil {
		if updateEvent.GetStatus().GetState() != mesos_v1.TaskState_TASK_FAILED {
			s.taskmanager.SetTaskLaunched(task)
		} else {
			s.taskmanager.Delete(task)
			id := task.TaskId.GetValue()
			if err := s.kv.Delete("/task/" + id); err != nil {
				s.logger.Emit(logging.ERROR, "Failed to delete task %s with name %s from persistent data store", id, task.GetName())
			}
		}
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

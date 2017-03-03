package eventcontroller

/*
Adapted from mesos-framework-sdk
*/
import (
	"fmt"
	"log"
	"mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/persistence"
	"mesos-framework-sdk/resources"
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/task_manager"
	"strconv"
)

type SprintEventController struct {
	scheduler       *scheduler.DefaultScheduler
	taskmanager     *task_manager.DefaultTaskManager
	resourcemanager *manager.DefaultResourceManager
	events          chan *sched.Event
	kv              persistence.KVStorage
}

func NewSprintEventController(scheduler *scheduler.DefaultScheduler, manager *task_manager.DefaultTaskManager, resourceManager *manager.DefaultResourceManager, eventChan chan *sched.Event, kv persistence.KVStorage) *SprintEventController {
	return &SprintEventController{
		taskmanager:     manager,
		scheduler:       scheduler,
		events:          eventChan,
		resourcemanager: resourceManager,
		kv:              kv,
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
	s.kv.Create("/frameworkId", idVal)
	log.Printf("Subscribed with an ID of %s", idVal)
}

func (s *SprintEventController) Run() {
	id, err := s.kv.Read("/frameworkId")
	if err == nil {
		s.scheduler.Info.Id = &mesos_v1.FrameworkID{Value: &id}
	}

	// TODO err is always nil here because of how Subscribe() is implemented
	err = s.scheduler.Subscribe(s.events)
	if err != nil {
		log.Printf("Failed to subscribe: %s", err.Error())
		return // TODO retry instead of returning
	}

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
				log.Println("Subscribe event.")
			case sched.Event_ERROR:
				go s.Error(t.GetError())
			case sched.Event_FAILURE:
				go s.Failure(t.GetFailure())
			case sched.Event_INVERSE_OFFERS:
				go s.InverseOffer(t.GetInverseOffers())
			case sched.Event_MESSAGE:
				go s.Message(t.GetMessage())
			case sched.Event_OFFERS:
				log.Println("Offers...")
				go s.Offers(t.GetOffers())
			case sched.Event_RESCIND:
				go s.Rescind(t.GetRescind())
			case sched.Event_RESCIND_INVERSE_OFFER:
				go s.RescindInverseOffer(t.GetRescindInverseOffer())
			case sched.Event_UPDATE:
				go s.Update(t.GetUpdate())
			case sched.Event_HEARTBEAT:
				fmt.Println("Heart beat.")
			case sched.Event_UNKNOWN:
				fmt.Println("Unknown event recieved.")
			}
		}
	}
}

func (s *SprintEventController) Offers(offerEvent *sched.Event_Offers) {
	fmt.Println("Offers event recieved.")
	//Reconcile any tasks.
	var reconcileTasks []*mesos_v1.Task
	s.scheduler.Reconcile(reconcileTasks)

	// Check task manager for any active tasks.
	if s.taskmanager.HasQueuedTasks() {
		// Update our resources in the manager
		s.resourcemanager.AddOffers(offerEvent.GetOffers())

		for _, mesosTask := range s.taskmanager.QueuedTasks() {
			// See if we have resources.
			if s.resourcemanager.HasResources() {
				offerIDs := []*mesos_v1.OfferID{}
				taskList := []*mesos_v1.TaskInfo{} // Clear it out every time.
				operations := []*mesos_v1.Offer_Operation{}

				offer, err := s.resourcemanager.Assign(mesosTask)
				if err != nil {
					// It didn't match any offers.
					log.Println(err.Error())
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

				taskList = append(taskList, t)
				offerIDs = append(offerIDs, offer.Id)
				operations = append(operations, resources.LaunchOfferOperation(taskList))

				log.Printf("Launching task %v\n", taskList)
				s.scheduler.Accept(offerIDs, operations, nil)

			}
		}
	} else {
		var ids []*mesos_v1.OfferID
		for _, v := range offerEvent.GetOffers() {
			ids = append(ids, v.GetId())
		}
		// decline offers.
		fmt.Println("Declining offers.")
		s.scheduler.Decline(ids, nil)
		s.scheduler.Suppress()
	}
}

func (s *SprintEventController) Rescind(rescindEvent *sched.Event_Rescind) {
	fmt.Printf("Rescind event recieved.: %v\n", *rescindEvent)
	rescindEvent.GetOfferId().GetValue()
}

func (s *SprintEventController) Update(updateEvent *sched.Event_Update) {
	fmt.Printf("Update recieved for: %v\n", *updateEvent.GetStatus())
	fmt.Printf("Network Info: %v\n", updateEvent.GetStatus().GetContainerStatus().GetNetworkInfos())

	task := s.taskmanager.Get(updateEvent.GetStatus().GetTaskId())
	// TODO: Handle more states in regard to tasks.
	if updateEvent.GetStatus().GetState() != mesos_v1.TaskState_TASK_FAILED {
		// Only set the task to "launched" if it didn't fail.
		s.taskmanager.SetTaskLaunched(task)
	} else {
		s.taskmanager.Delete(task)
	}
	status := updateEvent.GetStatus()
	s.scheduler.Acknowledge(status.GetAgentId(), status.GetTaskId(), status.GetUuid())
}

func (s *SprintEventController) Message(msg *sched.Event_Message) {
	fmt.Printf("Message event recieved: %v\n", *msg)
}

func (s *SprintEventController) Failure(fail *sched.Event_Failure) {
	log.Println("Executor " + fail.GetExecutorId().GetValue() + " failed with status " + strconv.Itoa(int(fail.GetStatus())))
}

func (s *SprintEventController) Error(err *sched.Event_Error) {
	fmt.Printf("Error event recieved: %v\n", err)
}

func (s *SprintEventController) InverseOffer(ioffers *sched.Event_InverseOffers) {
	fmt.Printf("Inverse Offer event recieved: %v\n", ioffers)
}

func (s *SprintEventController) RescindInverseOffer(rioffers *sched.Event_RescindInverseOffer) {
	fmt.Printf("Rescind Inverse Offer event recieved: %v\n", rioffers)
}

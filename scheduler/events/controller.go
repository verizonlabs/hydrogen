package events

/*
Adapted from mesos-framework-sdk
*/
import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/ha"
	"mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources"
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"net"
	"os"
	sprintSched "sprint/scheduler"
	"strconv"
	"time"
)

const (
	// Note (tim): Is there a reasonable non-linear equation to determine refuse seconds?
	// f^2/num_of_nodes_in_cluster where f is # of tasks to handle at once (per offer cycle).
	//
	refuseSeconds = 64.0
)

type SprintEventController struct {
	config          *sprintSched.Configuration
	scheduler       *scheduler.DefaultScheduler
	taskmanager     sdkTaskManager.TaskManager
	resourcemanager *manager.DefaultResourceManager
	events          chan *sched.Event
	kv              persistence.Storage
	logger          logging.Logger
	frameworkLease  *clientv3.LeaseID
	status          ha.Status
	name            string
}

// NOTE (tim): Cutting this signature down with newlines to make it easier to read.
// Please do this with any large signature to make it easier to parse.
func NewSprintEventController(
	config *sprintSched.Configuration,
	scheduler *scheduler.DefaultScheduler,
	manager sdkTaskManager.TaskManager,
	resourceManager *manager.DefaultResourceManager,
	eventChan chan *sched.Event,
	kv persistence.Storage,
	logger logging.Logger) *SprintEventController {

	return &SprintEventController{
		config:          config,
		taskmanager:     manager,
		scheduler:       scheduler,
		events:          eventChan,
		resourcemanager: resourceManager,
		kv:              kv,
		logger:          logger,
		status:          ha.Election,
		name:            utils.UuidAsString(),
	}
}

// This satisfies the Cluster Node Interface
func (s *SprintEventController) Name() (string, error) {
	return s.name, nil
}

func (s *SprintEventController) Status() (ha.Status, error) {
	return s.status, nil
}

// Implements our logic for either listening or leading.
func (s *SprintEventController) Communicate() {
	addr, err := net.ResolveTCPAddr(s.config.Leader.AddressFamily, "["+s.config.Leader.IP+"]:"+
		strconv.Itoa(s.config.Leader.ServerPort))
	if err != nil {
		s.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	// We start listening for connections.
	tcp, err := net.ListenTCP(s.config.Leader.AddressFamily, addr)
	if err != nil {
		s.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	s.status = ha.Listening

	for {
		// Block here until we get a new connection.
		// We don't want to do anything with the stream so move on without spawning a thread to handle the connection.
		conn, err := tcp.AcceptTCP()
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to accept client: %s", err.Error())
			time.Sleep(s.config.Leader.ServerRetry)
			continue
		}

		// TODO build out some config to use for setting the keep alive period here
		if err := conn.SetKeepAlive(true); err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set keep alive: %s", err.Error())
		}
	}

}

func (s *SprintEventController) Election() {
	for {
		// This will only set us as the leader if there isn't an already existing leader.
		s.CreateLeader()

		leader := s.GetLeader()
		if leader != s.config.Leader.IP {
			s.status = ha.Listening
			s.logger.Emit(logging.INFO, "Connecting to leader to determine when we need to wake up and perform leader election")

			// Block here until we lose connection to the leader.
			// Once the connection has been lost elect a new leader.
			err := s.leaderClient(leader)

			// Only delete the key if we've lost the connection, not timed out.
			// This conditional requires Go 1.6+

			// NOTE (tim): Casting here is dangerous and could lead to a panic
			if err, ok := err.(net.Error); ok && err.Timeout() {
				s.logger.Emit(logging.ERROR, "Timed out connecting to leader")
			} else {
				s.status = ha.Election
				s.logger.Emit(logging.ERROR, "Lost connection to leader")
				s.kv.Delete("/leader")
			}
		} else {
			s.status = ha.Leading
			s.logger.Emit(logging.INFO, "We're leading.")
			// We are the leader, exit the loop and start the scheduler/API.
			break
		}
	}
}

// Handles connections from other framework instances that try and determine the state of the leader.
// Used in coordination with determining if and when we need to perform leader election.
func (s *SprintEventController) leaderServer() {
	addr, err := net.ResolveTCPAddr(s.config.Leader.AddressFamily, "["+s.config.Leader.IP+"]:"+strconv.Itoa(s.config.Leader.ServerPort))
	if err != nil {
		s.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	tcp, err := net.ListenTCP(s.config.Leader.AddressFamily, addr)
	if err != nil {
		s.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	for {
		// Block here until we get a new connection.
		// We don't want to do anything with the stream so move on without spawning a thread to handle the connection.
		conn, err := tcp.AcceptTCP()
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to accept client: %s", err.Error())
			time.Sleep(s.config.Leader.ServerRetry)
			continue
		}

		// TODO build out some config to use for setting the keep alive period here
		if err := conn.SetKeepAlive(true); err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set keep alive: %s", err.Error())
		}
	}
}

// Connects to the leader and determines if and when we should start the leader election process.
func (s *SprintEventController) leaderClient(leader string) error {
	conn, err := net.DialTimeout(s.config.Leader.AddressFamily, "["+leader+"]:"+
		strconv.Itoa(s.config.Leader.ServerPort), 2*time.Second)
	if err != nil {
		return err
	}

	s.logger.Emit(logging.INFO, "Successfully connected to leader %s", leader)

	// TODO build out some config to use for setting the keep alive period here
	tcp := conn.(*net.TCPConn)
	if err := tcp.SetKeepAlive(true); err != nil {
		return err
	}

	s.status = ha.Talking
	buffer := make([]byte, 1)
	for {
		_, err := tcp.Read(buffer)
		if err != nil {
			return err
		}
	}
}

// Getter functions
func (s *SprintEventController) Scheduler() *scheduler.DefaultScheduler {
	return s.scheduler
}

// Getter function
func (s *SprintEventController) TaskManager() sdkTaskManager.TaskManager {
	return s.taskmanager
}

// Getter function
func (s *SprintEventController) ResourceManager() *manager.DefaultResourceManager {
	return s.resourcemanager
}

// Atomically create leader information.
func (s *SprintEventController) CreateLeader() {
	for {
		if err := s.kv.Create("/leader", s.config.Leader.IP); err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set leader information: "+err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		break
	}

}

// Atomically get leader information.
func (s *SprintEventController) GetLeader() string {
	for {
		leader, err := s.kv.Read("/leader")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get the leader: %s", err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}

		return leader[0]
	}
}

// Keep our state in check by periodically reconciling.
// This is recommended by Mesos.
func (s *SprintEventController) periodicReconcile() {
	ticker := time.NewTicker(s.config.Scheduler.ReconcileInterval)

	for {
		select {
		case <-ticker.C:

			recon, err := s.TaskManager().GetState(sdkTaskManager.RUNNING)
			if err != nil {
				// log here.
				continue
			}
			s.Scheduler().Reconcile(recon)
		}
	}
}

// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
func (s *SprintEventController) restoreTasks() {
	var tasks map[string]string
	var err error
	for {
		tasks, err = s.kv.Engine().(*etcd.Etcd).ReadAll("/tasks")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get all task data: %s", err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		break
	}

	for _, value := range tasks {
		var task sdkTaskManager.Task
		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			s.logger.Emit(logging.ERROR, err.Error())
		}

		var b bytes.Buffer
		b.Write(data)
		d := gob.NewDecoder(&b)
		err = d.Decode(&task)
		if err != nil {
			s.logger.Emit(logging.ERROR, err.Error())
		}
		s.TaskManager().Set(task.State, task.Info)
	}
}

// TODO think about renaming this to subscribed since the scheduler from the SDK is really handling the subscribe call.
func (s *SprintEventController) Subscribe(subEvent *sched.Event_Subscribed) {
	id := subEvent.GetFrameworkId()
	idVal := id.GetValue()
	s.scheduler.FrameworkInfo.Id = id
	s.logger.Emit(logging.INFO, "Subscribed with an ID of %s", idVal)

	// We pull the engine directly here to use non-interface methods.
	kv := s.kv.Engine().(*etcd.Etcd)

	var lease *clientv3.LeaseID
	var err error
	for {
		lease, err = kv.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.FrameworkInfo.GetFailoverTimeout()))
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		break
	}

	s.frameworkLease = lease

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
	// Start the election.
	s.logger.Emit(logging.INFO, "Starting leader election socket server")
	go s.Communicate()

	// Block here until we either become a leader or a standby.
	// If we are the leader we break out and continue to execute the rest of the scheduler.
	// If we are a standby then we connect to the leader and wait for the process to start over again.
	s.Election()

	for {
		id, err := s.kv.Read("/frameworkId")
		if err == nil {
			s.scheduler.FrameworkInfo.Id = &mesos_v1.FrameworkID{Value: &id[0]}
			break
		} else {
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
	}

	// Recover our state (if any) in the event we (or the server) go down.
	s.logger.Emit(logging.INFO, "Restoring any persisted state from data store")
	s.restoreTasks()

	// Kick off our scheduled reconciling.
	s.logger.Emit(logging.INFO, "Starting periodic reconciler thread with a %g minute interval", s.config.Scheduler.ReconcileInterval.Minutes())
	go s.periodicReconcile()

	go func() {
		for {
			var leader []string
			var err error
			for {
				leader, err = s.kv.Read("/leader")
				if err != nil {
					s.logger.Emit(logging.ERROR, "Failed to find the leader: %s", err.Error())
					time.Sleep(s.config.Persistence.RetryInterval)
					continue
				}
				break
			}

			// We should only ever reach here if we hit a network partition and the standbys lose connection to the leader.
			// If this happens we need to check if there really is another leader alive that we just can't reach.
			// If we wrongly think we are the leader and try to subscribe when there's already a leader then we will disconnect the leader.
			// Both the leader and the incorrectly determined new leader will continue to disconnect each other.
			if leader[0] != s.config.Leader.IP {
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
				s.Subscribe(t.GetSubscribed())
			case sched.Event_ERROR:
				s.Error(t.GetError())
			case sched.Event_FAILURE:
				s.Failure(t.GetFailure())
			case sched.Event_INVERSE_OFFERS:
				s.InverseOffer(t.GetInverseOffers())
			case sched.Event_MESSAGE:
				s.Message(t.GetMessage())
			case sched.Event_OFFERS:
				s.Offers(t.GetOffers())
			case sched.Event_RESCIND:
				s.Rescind(t.GetRescind())
			case sched.Event_RESCIND_INVERSE_OFFER:
				s.RescindInverseOffer(t.GetRescindInverseOffer())
			case sched.Event_UPDATE:
				s.Update(t.GetUpdate())
			case sched.Event_HEARTBEAT:
				for {
					if err := s.kv.Engine().(*etcd.Etcd).RefreshLease(s.frameworkLease); err != nil {
						s.logger.Emit(logging.ERROR, "Failed to refresh framework ID lease: %s", err.Error())
						time.Sleep(s.config.Persistence.RetryInterval)
						continue
					}
					break
				}
			case sched.Event_UNKNOWN:
				s.logger.Emit(logging.ALARM, "Unknown event received")
			}
		}
	}
}

func (s *SprintEventController) declineOffers(offers []*mesos_v1.Offer, refuseSeconds float64) {
	declineIDs := []*mesos_v1.OfferID{}
	// Decline whatever offers are left over
	for _, id := range offers {
		declineIDs = append(declineIDs, id.GetId())
	}

	s.scheduler.Decline(declineIDs, &mesos_v1.Filters{RefuseSeconds: proto.Float64(refuseSeconds)})
}

func (s *SprintEventController) Offers(offerEvent *sched.Event_Offers) {
	// Check if we have any in the task manager we want to launch
	queued, err := s.taskmanager.GetState(sdkTaskManager.UNKNOWN)

	if err != nil {
		s.logger.Emit(logging.INFO, "No tasks to launch.")
		s.scheduler.Suppress()

		s.declineOffers(offerEvent.GetOffers(), refuseSeconds) // All offers to decline.
		return
	}

	// Update our resources in the manager
	s.resourcemanager.AddOffers(offerEvent.GetOffers())
	accepts := make(map[*mesos_v1.OfferID][]*mesos_v1.Offer_Operation)

	for _, mesosTask := range queued {
		if !s.resourcemanager.HasResources() {
			break
		}

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

		// TODO (aaron) investigate this state further as it might cause side effects.
		// this is artificially set to STAGING, it does not correspond to when Mesos sets this task as STAGING.
		// for example other parts of the codebase may check for STAGING and this would cause it to be set too early.
		s.TaskManager().Set(sdkTaskManager.STAGING, t)
		accepts[offer.Id] = append(accepts[offer.Id], resources.LaunchOfferOperation([]*mesos_v1.TaskInfo{t}))
	}

	// Multiplex our tasks onto as few offers as possible and launch them all.
	for id, launches := range accepts {

		// TODO (tim) The offer operations will need to be parsed for volume mounting and etc.)
		s.scheduler.Accept([]*mesos_v1.OfferID{id}, launches, nil)
	}

	// Resource manager pops offers when they are accepted
	// Offers() returns a list of what is left, therefore whatever is left is to be rejected.
	s.declineOffers(s.ResourceManager().Offers(), refuseSeconds)
}

func (s *SprintEventController) Rescind(rescindEvent *sched.Event_Rescind) {
	s.logger.Emit(logging.INFO, "Rescind event recieved: %v", *rescindEvent)
	rescindEvent.GetOfferId().GetValue()
}

func (s *SprintEventController) Update(updateEvent *sched.Event_Update) {
	status := updateEvent.GetStatus()
	taskId := status.GetTaskId()
	task, err := s.taskmanager.GetById(taskId)
	if err != nil {
		// The event is from a task that has been deleted from the task manager,
		// ignore updates.
		// NOTE (tim): Do we want to keep deleted task history for a certain amount of time
		// before it's deleted? We would record status updates after it's killed here.
		// ACK update, return.
		status := updateEvent.GetStatus()
		s.scheduler.Acknowledge(status.GetAgentId(), status.GetTaskId(), status.GetUuid())
		return
	}

	state := status.GetState()
	message := status.GetMessage()
	agentId := status.GetAgentId()
	s.taskmanager.Set(state, task)

	switch state {
	case mesos_v1.TaskState_TASK_FAILED:
		// TODO (tim): Check task manager for task retry policy, then retry as given.
		// Default for now is just retry forever.
		s.logger.Emit(logging.ERROR, message)
		s.taskmanager.Set(sdkTaskManager.UNKNOWN, task)
	case mesos_v1.TaskState_TASK_STAGING:
		// NOP, keep task set to "launched".
	case mesos_v1.TaskState_TASK_DROPPED:
		// Transient error, we should retry launching. Taskinfo is fine.
	case mesos_v1.TaskState_TASK_ERROR:
		// TODO (tim): Error with the taskinfo sent to the agent. Give verbose reasoning back.
	case mesos_v1.TaskState_TASK_FINISHED:
		s.taskmanager.Delete(task)
	case mesos_v1.TaskState_TASK_GONE:
		// Agent is dead and task is lost.
	case mesos_v1.TaskState_TASK_GONE_BY_OPERATOR:
		// Agent might be dead, master is unsure. Will return to RUNNING state possibly or die.
	case mesos_v1.TaskState_TASK_KILLED:
		// Task was killed.
		s.logger.Emit(
			logging.INFO,
			"Task %s on agent %s was killed",
			taskId.GetValue(),
			agentId.GetValue(),
		)
		s.taskmanager.Delete(task)
	case mesos_v1.TaskState_TASK_KILLING:
		// Task is in the process of catching a SIGNAL and shutting down.
	case mesos_v1.TaskState_TASK_LOST:
		// Task is unknown to the master and lost. Should reschedule.
		s.logger.Emit(logging.ALARM, "Task %s was lost", taskId.GetValue())
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

	s.scheduler.Acknowledge(agentId, taskId, status.GetUuid())
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

package executor

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/pborman/uuid"
	"mesos-sdk/executor"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	"mesos-sdk/executor/calls"
	"mesos-sdk/executor/events"
	"mesos-sdk/httpcli"
	"mesos-sdk"
	"io"
	"log"
	"time"
)

/*
	Executor Struct holds all information for top-level executor to function
*/
type SprintExecutor struct {
	Framework      mesos.FrameworkInfo
	Executor       mesos.ExecutorInfo
	Agent          mesos.AgentInfo
	config         *Configuration
	http           *httpcli.Client
	options        executor.CallOptions
	unackedTasks   map[mesos.TaskID]mesos.TaskInfo
	unackedUpdates map[string]executor.Call_Update
	failedTasks    map[mesos.TaskID]mesos.TaskStatus
	subscribe      *executor.Call
	reconnect      <-chan struct{}
	disconnected   time.Time
	eventHandler   events.Handler
	shutdown       bool
}

/*
	NewExecutor returns a new executor according to the configuration passed to it.
*/
func NewExecutor(cfg *Configuration) *SprintExecutor {
	return &SprintExecutor{
		config: cfg,
		http: httpcli.New(
			httpcli.Endpoint(cfg.ExecutorConfig.AgentEndpoint),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.Timeout),
				),
			),
		),
		options: executor.CallOptions{
			calls.Framework(cfg.ExecutorConfig.FrameworkID),
			calls.Executor(cfg.ExecutorConfig.ExecutorID),
		},
		unackedTasks:   make(map[mesos.TaskID]mesos.TaskInfo),
		unackedUpdates: make(map[string]executor.Call_Update),
		failedTasks:    make(map[mesos.TaskID]mesos.TaskStatus),
		reconnect:      backoff.Notifier(10*time.Second, cfg.ExecutorConfig.SubscriptionBackoffMax*4, nil),
		disconnected:   time.Now(),
	}
}

/*
	Default Handlers passed in here. Otherwise, pass in a struct that statisfies the
 */
func (s *SprintExecutor) NewExecutorHandler() events.Handler {
	return s.build()
}

func (s *SprintExecutor) Subscribe(){
	s.subscribe = calls.Subscribe(nil, nil).With(s.options...)
}

func (s *SprintExecutor) build() *events.Mux {
	Mux := events.NewMux(
		s.defaultSubscribeHandler(),
		s.defaultAcknowledgeHandler(),
		s.defaultKillHandler(),
		s.defaultErrorHandler(),
		s.defaultLaunchHandler(),
		s.defaultMessageHandler(),
		s.defaultShutdownHandler(),
	)
	return Mux
}

/*
	Run starts the main loop that will poll for events.
*/
func (s *SprintExecutor) Run() {
	for {
		s.subscribe = s.subscribe.With(
			s.unacknowledgedTasks(),
			s.unacknowledgedUpdates(),
		)
		// Anon func to avoid defer memory leak.
		func() {
			resp, err := s.http.Do(s.subscribe, httpcli.Close(true))

			if resp != nil {
				defer resp.Close()
			}

			if err == nil {
				err = s.eventLoop(resp.Decoder())
				s.disconnected = time.Now()
			}
			if err != nil && err != io.EOF {
				log.Println(err)
			} else {
				log.Println("disconnected")
			}
		}()
		if s.shutdown {
			return
		}
		if !s.config.ExecutorConfig.Checkpoint {
			return
		}
		if time.Now().Sub(s.disconnected) > s.config.ExecutorConfig.RecoveryTimeout {
			return
		}
		<-s.reconnect

	}
}

func (s *SprintExecutor) unacknowledgedTasks() executor.CallOpt {
	return func(call *executor.Call) {
		if n := len(s.unackedTasks); n > 0 {
			unackedTasks := make([]mesos.TaskInfo, 0, n)
			for k := range s.unackedTasks {
				unackedTasks = append(unackedTasks, s.unackedTasks[k])
			}
			call.Subscribe.UnacknowledgedTasks = unackedTasks
		} else {
			call.Subscribe.UnacknowledgedTasks = nil
		}
	}
}

func (s *SprintExecutor) unacknowledgedUpdates() executor.CallOpt {
	return func(call *executor.Call) {
		if n := len(s.unackedUpdates); n > 0 {
			unackedUpdates := make([]executor.Call_Update, 0, n)
			for k := range s.unackedUpdates {
				unackedUpdates = append(unackedUpdates, s.unackedUpdates[k])
			}
			call.Subscribe.UnacknowledgedUpdates = unackedUpdates
		} else {
			call.Subscribe.UnacknowledgedUpdates = nil
		}
	}
}

func (s *SprintExecutor) eventLoop(decoder encoding.Decoder) (err error) {
	for err == nil && !s.shutdown {
		s.sendfailedTasks()

		var event executor.Event
		if err = decoder.Invoke(&event); err == nil {
			err = s.eventHandler.HandleEvent(&event)
		}
	}
	return err
}

func (s *SprintExecutor) sendfailedTasks() {
	for taskID, status := range s.failedTasks {
		updateErr := s.update(status)
		if updateErr != nil {
			log.Printf("failed to send status update for task %s: %v", taskID.Value, updateErr)
		} else {
			delete(s.failedTasks, taskID)
		}
	}
}

func (s *SprintExecutor) launch(task mesos.TaskInfo) {
	s.unackedTasks[task.TaskID] = task

	status := s.newStatus(task.TaskID)
	status.State = mesos.TASK_RUNNING.Enum()
	err := s.update(status)
	if err != nil {
		log.Printf("failed to send TASK_RUNNING for task %s: %+v", task.TaskID.Value, err)
		status.State = mesos.TASK_FAILED.Enum()
		status.Message = proto.String(err.Error())
		s.failedTasks[task.TaskID] = status
		return
	}

	status = s.newStatus(task.TaskID)
	status.State = mesos.TASK_FINISHED.Enum()
	err = s.update(status)
	if err != nil {
		log.Printf("failed to send TASK_FINISHED for task %s: %+v", task.TaskID.Value, err)
		status.State = mesos.TASK_FAILED.Enum()
		status.Message = proto.String(err.Error())
		s.failedTasks[task.TaskID] = status
	}
}

func (s *SprintExecutor) update(status mesos.TaskStatus) error {
	updatedState := calls.Update(status).With(s.options...)
	resp, err := s.http.Do(updatedState)
	if resp != nil {
		resp.Close()
	}
	if err != nil {
		log.Printf("failed to send update: %v", err)
	} else {
		s.unackedUpdates[string(status.UUID)] = *updatedState.Update
	}
	return err
}

func (s *SprintExecutor) newStatus(id mesos.TaskID) mesos.TaskStatus {
	return mesos.TaskStatus{
		TaskID:     id,
		Source:     mesos.SOURCE_EXECUTOR.Enum(),
		ExecutorID: &s.Executor.ExecutorID,
		UUID:       []byte(uuid.NewRandom()),
	}
}

/*
This is the first step in the communication process between the executor and agent.
This is also to be considered as subscription to the “/executor” events stream.
To subscribe with the agent, the executor sends a HTTP POST request with encoded SUBSCRIBE messags.
The HTTP response is a stream with RecordIO encoding, with the first event being SUBSCRIBED events.
*/
func (s *SprintExecutor) defaultSubscribeHandler() events.Option {
	return events.Handle(executor.Event_SUBSCRIBED, events.HandlerFunc(func(e *executor.Event) error {
		log.Print("Executor has subscribed.")
		s.Framework = e.Subscribed.FrameworkInfo
		s.Executor = e.Subscribed.ExecutorInfo
		s.Agent = e.Subscribed.AgentInfo
		return nil
	}))
}

/*
Sent by the agent whenever it needs to assign a new task to the executor.
*/
func (s *SprintExecutor) defaultLaunchHandler() events.Option {
	return events.Handle(executor.Event_LAUNCH, events.HandlerFunc(func(e *executor.Event) error {
		log.Print("Executor is Launching.")
		s.launch(e.Launch.Task)
		return nil
	}))
}

/*
The KILL event is sent whenever the scheduler needs to stop execution of a specific task.
The executor is required to send a terminal update (s.g., TASK_FINISHED, TASK_KILLED or TASK_FAILED) back to the agent
once it has stopped/killed the task. Mesos will mark the task resources as freed once the terminal update is received.
*/
func (s *SprintExecutor) defaultKillHandler() events.Option {
	return events.Handle(executor.Event_KILL, events.HandlerFunc(func(e *executor.Event) error {
		log.Print("Executor is killing a task.")
		return nil
	}))
}

/*
Sent by the agent in order to signal the executor that a status update was received as part of the reliable message
passing mechanism. Acknowledged updates must not be retried.
*/
func (s *SprintExecutor) defaultAcknowledgeHandler() events.Option {
	return events.Handle(executor.Event_ACKNOWLEDGED, events.HandlerFunc(func(e *executor.Event) error {
		log.Print("Executor is Acknowledging.")
		delete(s.unackedTasks, e.Acknowledged.TaskID)
		delete(s.unackedUpdates, string(e.Acknowledged.UUID))
		return nil
	}))
}

/*
Custom message generated by the scheduler and forwarded all the way to the executor. These messages are delivered
“as-is” by Mesos and have no delivery guarantees. It is up to the scheduler to retry if a message is dropped for
any reason. Note that data is raw bytes encoded as Base64.
*/
func (s *SprintExecutor) defaultMessageHandler() events.Option {
	return events.Handle(executor.Event_MESSAGE, events.HandlerFunc(func(e *executor.Event) error {
		log.Printf("MESSAGE: received %d bytes of message data", len(e.Message.Data))
		return nil
	}))
}

/*
Sent by the agent in order to shutdown the executor. Once an executor gets a SHUTDOWN event it is required to kill all
its tasks, send TASK_KILLED updates and gracefully exit. If an executor doesn’t terminate within a certain period
MESOS_EXECUTOR_SHUTDOWN_GRACE_PERIOD (an environment variable set by the agent upon executor startup), the agent will
forcefully destroy the container where the executor is running. The agent would then send TASK_LOST updates for any
remaining active tasks of this executor.
*/
func (s *SprintExecutor) defaultShutdownHandler() events.Option {
	return events.Handle(executor.Event_SHUTDOWN, events.HandlerFunc(func(e *executor.Event) error {
		log.Print("Executor is shutting down...")
		s.shutdown = true
		return nil
	}))
}

/*
Sent by the agent when an asynchronous error event is generated. It is recommended that the executor abort when it
receives an error event and retry subscription.
*/
func (s *SprintExecutor) defaultErrorHandler() events.Option {
	return events.Handle(executor.Event_ERROR, events.HandlerFunc(func(e *executor.Event) error {
		log.Print("Executor encountered an error...")
		return errors.New("received abort signal from mesos, will attempt to re-subscribe")
	}))
}

/*
   Interface for Executor Events
*/
type ExecutorHandler interface {
	Subscribed(events.HandlerFunc) events.Option
	Launch(events.HandlerFunc) events.Option
	Kill(events.HandlerFunc) events.Option
	Acknowledged(events.HandlerFunc) events.Option
	Message(events.HandlerFunc) events.Option
	Shutdown(events.HandlerFunc) events.Option
	Error(events.HandlerFunc) events.Option
}

// Function pointer
type eventHandler func(int32, events.HandlerFunc) error

type ExecutorHandlers struct {
	Mux *events.Mux
}

// Used to add a handler to Mux.
func (e *ExecutorHandlers) addHandler(event executor.Event_Type, handlerFunc events.HandlerFunc) events.Option {
	option := events.Handle(event, handlerFunc)
	return option
}

func (e *ExecutorHandlers) Subscribed(handlerFunc events.HandlerFunc) events.Option {
	return e.addHandler(executor.Event_SUBSCRIBED, handlerFunc)
}

func (e *ExecutorHandlers) Launch(handlerFunc events.HandlerFunc) events.Option {
	return e.addHandler(executor.Event_LAUNCH, handlerFunc)
}

func (e *ExecutorHandlers) Kill(handlerFunc events.HandlerFunc) events.Option {
	return e.addHandler(executor.Event_KILL, handlerFunc)
}

func (e *ExecutorHandlers) Acknowledged(handlerFunc events.HandlerFunc) events.Option {
	return e.addHandler(executor.Event_ACKNOWLEDGED, handlerFunc)
}

func (e *ExecutorHandlers) Message(handlerFunc events.HandlerFunc) events.Option {
	return e.addHandler(executor.Event_MESSAGE, handlerFunc)
}

func (e *ExecutorHandlers) Shutdown(handlerFunc events.HandlerFunc) events.Option {
	return e.addHandler(executor.Event_SHUTDOWN, handlerFunc)
}

func (e *ExecutorHandlers) Error(handlerFunc events.HandlerFunc) events.Option {
	return e.addHandler(executor.Event_ERROR, handlerFunc)
}

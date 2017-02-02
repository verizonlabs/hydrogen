package executor

import (
	"errors"
	"io"
	"log"
	"mesos-sdk"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	exec "mesos-sdk/executor"
	"mesos-sdk/executor/calls"
	"mesos-sdk/executor/events"
	"mesos-sdk/extras"
	"mesos-sdk/httpcli"
	"os"
	"time"
)

// Base implementation of an executor.
type executor interface {
	Run()
}

// Holds all necessary information for our executor to function.
type sprintExecutor struct {
	callOptions    exec.CallOptions
	cli            *httpcli.Client
	config         configuration
	framework      mesos.FrameworkInfo
	executor       mesos.ExecutorInfo
	agent          mesos.AgentInfo
	handler        events.Handler
	unackedTasks   map[mesos.TaskID]mesos.TaskInfo
	unackedUpdates map[string]exec.Call_Update
	failedTasks    map[mesos.TaskID]mesos.TaskStatus
	shouldQuit     bool
}

// Returns a new executor using user-supplied configuration.
func NewExecutor(cfg configuration) *sprintExecutor {
	executor := sprintExecutor{
		callOptions: exec.CallOptions{
			calls.Framework(os.Getenv("MESOS_FRAMEWORK_ID")),
			calls.Executor(os.Getenv("MESOS_EXECUTOR_ID")),
		},
		cli: httpcli.New(
			httpcli.Endpoint(cfg.Endpoint()),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.Timeout()),
				),
			),
		),
		config:         cfg,
		unackedTasks:   make(map[mesos.TaskID]mesos.TaskInfo),
		unackedUpdates: make(map[string]exec.Call_Update),
		failedTasks:    make(map[mesos.TaskID]mesos.TaskStatus),
	}
	executor.handler = executor.buildEventHandler()

	return &executor
}

// Start execution of the executor
func (e *sprintExecutor) Run() {
	shouldReconnect := backoff.Notifier(1*time.Second, e.config.SubscriptionBackoffMax()*4/3, nil)
	disconnected := time.Now()
	subscribe := calls.Subscribe(nil, nil).With(e.callOptions...)

	for {
		subscribe = subscribe.With(
			e.unacknowledgedTasks(),
			e.unacknowledgedUpdates(),
		)

		resp, err := e.cli.Do(subscribe, httpcli.Close(true))
		if err == nil {
			// we're officially connected, start decoding events
			err = e.eventLoop(resp.Decoder())
			resp.Close()
			disconnected = time.Now()
		}
		// If the call/connection fails, we get here.
		if err != nil && err != io.EOF {
			log.Println(err)
		} else {
			log.Println("Executor disconnected")
		}

		if e.shouldQuit {
			log.Println("Gracefully shutting down...")
			return
		}
		if !e.config.Checkpoint() {
			log.Println("Gracefully exiting because framework checkpointing is not enabled")
			return
		}
		if time.Now().Sub(disconnected) > e.config.RecoveryTimeout() {
			log.Printf("Failed to re-establish subscription with agent within %v, aborting", e.config.RecoveryTimeout())
			return
		}

		<-shouldReconnect // Wait for some amount of time before retrying subscription
	}
}

// UnacknowledgedTasks is a functional option that sets the value of the UnacknowledgedTasks field of a Subscribe call.
func (e *sprintExecutor) unacknowledgedTasks() exec.CallOpt {
	return func(call *exec.Call) {
		if n := len(e.unackedTasks); n > 0 {
			unackedTasks := make([]mesos.TaskInfo, 0, n)
			for k := range e.unackedTasks {
				unackedTasks = append(unackedTasks, e.unackedTasks[k])
			}
			call.Subscribe.UnacknowledgedTasks = unackedTasks
		} else {
			call.Subscribe.UnacknowledgedTasks = nil
		}
	}
}

// UnacknowledgedUpdates is a functional option that sets the value of the UnacknowledgedUpdates field of a Subscribe call.
func (e *sprintExecutor) unacknowledgedUpdates() exec.CallOpt {
	return func(call *exec.Call) {
		if n := len(e.unackedUpdates); n > 0 {
			unackedUpdates := make([]exec.Call_Update, 0, n)
			for k := range e.unackedUpdates {
				unackedUpdates = append(unackedUpdates, e.unackedUpdates[k])
			}
			call.Subscribe.UnacknowledgedUpdates = unackedUpdates
		} else {
			call.Subscribe.UnacknowledgedUpdates = nil
		}
	}
}

// Processes and decodes events from our Mesos stream.
func (e *sprintExecutor) eventLoop(decoder encoding.Decoder) (err error) {
	for err == nil && !e.shouldQuit {
		e.sendFailedTasks()

		var event exec.Event
		if err = decoder.Invoke(&event); err == nil {
			err = e.handler.HandleEvent(&event)
		}
	}
	return err
}

// Defines event handlers for various events that will be received from Mesos.
func (e *sprintExecutor) buildEventHandler() events.Handler {
	return events.NewMux(
		events.Handle(exec.Event_SUBSCRIBED, events.HandlerFunc(func(event *exec.Event) error {
			log.Println("SUBSCRIBED")
			e.framework = event.Subscribed.FrameworkInfo
			e.executor = event.Subscribed.ExecutorInfo
			e.agent = event.Subscribed.AgentInfo
			return nil
		})),
		events.Handle(exec.Event_LAUNCH, events.HandlerFunc(func(event *exec.Event) error {
			e.launch(event.Launch.Task)
			return nil
		})),
		events.Handle(exec.Event_KILL, events.HandlerFunc(func(event *exec.Event) error {
			// TODO implement this
			log.Println("warning: KILL not implemented")
			return nil
		})),
		events.Handle(exec.Event_ACKNOWLEDGED, events.HandlerFunc(func(event *exec.Event) error {
			delete(e.unackedTasks, event.Acknowledged.TaskID)
			delete(e.unackedUpdates, string(event.Acknowledged.UUID))
			return nil
		})),
		events.Handle(exec.Event_MESSAGE, events.HandlerFunc(func(event *exec.Event) error {
			log.Printf("MESSAGE: received %d bytes of message data", len(event.Message.Data))
			return nil
		})),
		events.Handle(exec.Event_SHUTDOWN, events.HandlerFunc(func(event *exec.Event) error {
			log.Println("SHUTDOWN received")
			e.shouldQuit = true
			return nil
		})),
		events.Handle(exec.Event_ERROR, events.HandlerFunc(func(event *exec.Event) error {
			// TODO implement this
			log.Println("ERROR received")
			return errors.New("received abort signal from mesos, will attempt to re-subscribe")
		})),
	)
}

// Notify Mesos of any failed tasks we have.
func (e *sprintExecutor) sendFailedTasks() {
	for taskID, status := range e.failedTasks {
		updateErr := e.update(status)
		if updateErr != nil {
			log.Printf("failed to send status update for task %s: %+v", taskID.Value, updateErr)
		} else {
			delete(e.failedTasks, taskID)
		}
	}
}

// Tell Mesos that we're starting or finishing our tasks.
func (e *sprintExecutor) launch(task mesos.TaskInfo) {
	e.unackedTasks[task.TaskID] = task

	// send RUNNING
	status := e.newStatus(task.TaskID, mesos.TASK_RUNNING)
	err := e.update(status)
	if err != nil {
		log.Printf("failed to send TASK_RUNNING for task %s: %+v", task.TaskID.Value, err)
		status.State = mesos.TASK_FAILED.Enum()
		status.Message = extras.ProtoString(err.Error())
		e.failedTasks[task.TaskID] = status
		return
	}

	// send FINISHED
	status = e.newStatus(task.TaskID, mesos.TASK_FINISHED)
	err = e.update(status)
	if err != nil {
		log.Printf("failed to send TASK_FINISHED for task %s: %+v", task.TaskID.Value, err)
		status.State = mesos.TASK_FAILED.Enum()
		status.Message = extras.ProtoString(err.Error())
		e.failedTasks[task.TaskID] = status
	}
}

// Sends a status update to Mesos.
func (e *sprintExecutor) update(status mesos.TaskStatus) error {
	upd := calls.Update(status).With(e.callOptions...)
	resp, err := e.cli.Do(upd)
	if resp != nil {
		resp.Close()
	}
	if err != nil {
		log.Printf("failed to send update: %+v", err)
	} else {
		e.unackedUpdates[string(status.UUID)] = *upd.Update
	}
	return err
}

// Constructs a new TaskStatus which can be used to send Mesos status updates.
func (e *sprintExecutor) newStatus(id mesos.TaskID, state mesos.TaskState) mesos.TaskStatus {
	return mesos.TaskStatus{
		TaskID:     id,
		Source:     mesos.SOURCE_EXECUTOR.Enum(),
		ExecutorID: &e.executor.ExecutorID,
		UUID:       extras.Uuid(),
		State:      state.Enum(),
	}
}

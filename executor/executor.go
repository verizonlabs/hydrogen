package executor

import (
	"errors"
	"io"
	"log"
	"mesos-sdk"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	"mesos-sdk/executor"
	"mesos-sdk/extras"
	"mesos-sdk/executor/calls"
	"mesos-sdk/executor/config"
	"mesos-sdk/executor/events"
	"mesos-sdk/httpcli"
	"net/url"
	"time"
)

const (
	apiPath     = "/api/v1/executor"
	httpTimeout = 10 * time.Second

	// Fetcher for the executor
	executorFetcherPath  = "/executor"
	executorFetcherPort  = "8081"
	executorFetcherHost  = "localhost"
	executorFetcherProto = "http"
	executorFetcherURI   = executorFetcherProto + "://" + executorFetcherHost + ":" + executorFetcherPort + executorFetcherPath
	isExecutable         = true
)

var errMustAbort = errors.New("received abort signal from mesos, will attempt to re-subscribe")

func stringToStringPointer(str string) *string {
	var ptr = new(string)
	*ptr = str
	return ptr
}

func boolToBoolPointer(b bool) *bool {
	var ptr = new(bool)
	*ptr = b
	return ptr
}

// TODO make a NewExecutorWithConfig()
func NewExecutor() *mesos.ExecutorInfo {
	return &mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{Value: extras.Uuid()},
		Name:       stringToStringPointer("Sprinter"),
		Command:    getCommandInfo("echo 'hello world'"),
		Resources: []mesos.Resource{
			getCpuResources(0.5),
			getMemResources(1024.0),
		},
		Container: getContainer("busybox:latest"),
	}
}

/*
	Returns a commandInfo protobuf with the specified command
*/
func getCommandInfo(command string) mesos.CommandInfo {
	return mesos.CommandInfo{
		Value: stringToStringPointer(command),
		// URI is needed to pull the executor down!
		URIs: []mesos.CommandInfo_URI{
			{
				Value:      executorFetcherURI,
				Executable: boolToBoolPointer(isExecutable),
			},
		},
	}
}

/*
	Convenience function to get a mesos container running a docker image.
	Requires slave to run with mesos containerization flag set.
*/
func getContainer(imageName string) *mesos.ContainerInfo {
	return &mesos.ContainerInfo{
		Type: mesos.ContainerInfo_MESOS.Enum(),
		Mesos: &mesos.ContainerInfo_MesosInfo{
			Image: &mesos.Image{
				Docker: &mesos.Image_Docker{
					Name: imageName,
				},
				Type: mesos.Image_DOCKER.Enum(),
			},
		},
	}
}

/*
	Convenience function to return mesos CPU resources.
	CPU resources can be fractional.
*/
func getCpuResources(amount float64) mesos.Resource {
	return mesos.Resource{
		Name:   "cpus",
		Type:   mesos.SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: 0.5},
	}
}

/*
	Convenience function to return mesos memory resources.
	Memory allocated is in mb.
*/
func getMemResources(amount float64) mesos.Resource {
	return mesos.Resource{
		Name:   "mem",
		Type:   mesos.SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: amount},
	}
}

func Run(cfg config.Config) {
	var (
		apiURL = url.URL{
			Scheme: "http",
			Host:   cfg.AgentEndpoint,
			Path:   apiPath,
		}
		state = &executorState{
			cli: httpcli.New(
				httpcli.Endpoint(apiURL.String()),
				httpcli.Codec(&encoding.ProtobufCodec),
				httpcli.Do(httpcli.With(httpcli.Timeout(httpTimeout))),
			),
			callOptions: executor.CallOptions{
				calls.Framework(cfg.FrameworkID),
				calls.Executor(cfg.ExecutorID),
			},
			unackedTasks:   make(map[mesos.TaskID]mesos.TaskInfo),
			unackedUpdates: make(map[string]executor.Call_Update),
			failedTasks:    make(map[mesos.TaskID]mesos.TaskStatus),
		}
		subscribe       = calls.Subscribe(nil, nil).With(state.callOptions...)
		shouldReconnect = backoff.Notifier(1*time.Second, cfg.SubscriptionBackoffMax*4/3, nil)
		disconnected    = time.Now()
		handler         = buildEventHandler(state)
	)
	for {
		subscribe = subscribe.With(
			unacknowledgedTasks(state),
			unacknowledgedUpdates(state),
		)
		func() {
			resp, err := state.cli.Do(subscribe, httpcli.Close(true))
			if resp != nil {
				defer resp.Close()
			}
			if err == nil {
				// we're officially connected, start decoding events
				err = eventLoop(state, resp.Decoder(), handler)
				disconnected = time.Now()
			}
			if err != nil && err != io.EOF {
				log.Println(err)
			} else {
				log.Println("disconnected")
			}
		}()
		if state.shouldQuit {
			log.Println("gracefully shutting down because we were told to")
			return
		}
		if !cfg.Checkpoint {
			log.Println("gracefully exiting because framework checkpointing is NOT enabled")
			return
		}
		if time.Now().Sub(disconnected) > cfg.RecoveryTimeout {
			log.Printf("failed to re-establish subscription with agent within %v, aborting", cfg.RecoveryTimeout)
			return
		}
		<-shouldReconnect // wait for some amount of time before retrying subscription
	}
}

// unacknowledgedTasks is a functional option that sets the value of the UnacknowledgedTasks
// field of a Subscribe call.
func unacknowledgedTasks(state *executorState) executor.CallOpt {
	return func(call *executor.Call) {
		if n := len(state.unackedTasks); n > 0 {
			unackedTasks := make([]mesos.TaskInfo, 0, n)
			for k := range state.unackedTasks {
				unackedTasks = append(unackedTasks, state.unackedTasks[k])
			}
			call.Subscribe.UnacknowledgedTasks = unackedTasks
		} else {
			call.Subscribe.UnacknowledgedTasks = nil
		}
	}
}

// unacknowledgedUpdates is a functional option that sets the value of the UnacknowledgedUpdates
// field of a Subscribe call.
func unacknowledgedUpdates(state *executorState) executor.CallOpt {
	return func(call *executor.Call) {
		if n := len(state.unackedUpdates); n > 0 {
			unackedUpdates := make([]executor.Call_Update, 0, n)
			for k := range state.unackedUpdates {
				unackedUpdates = append(unackedUpdates, state.unackedUpdates[k])
			}
			call.Subscribe.UnacknowledgedUpdates = unackedUpdates
		} else {
			call.Subscribe.UnacknowledgedUpdates = nil
		}
	}
}

func eventLoop(state *executorState, decoder encoding.Decoder, h events.Handler) (err error) {
	for err == nil && !state.shouldQuit {
		sendFailedTasks(state)

		var e executor.Event
		if err = decoder.Invoke(&e); err == nil {
			err = h.HandleEvent(&e)
		}
	}
	return err
}

func buildEventHandler(state *executorState) events.Handler {
	return events.NewMux(
		events.Handle(executor.Event_SUBSCRIBED, events.HandlerFunc(func(e *executor.Event) error {
			log.Println("SUBSCRIBED")
			state.framework = e.Subscribed.FrameworkInfo
			state.executor = e.Subscribed.ExecutorInfo
			state.agent = e.Subscribed.AgentInfo
			return nil
		})),
		events.Handle(executor.Event_LAUNCH, events.HandlerFunc(func(e *executor.Event) error {
			launch(state, e.Launch.Task)
			return nil
		})),
		events.Handle(executor.Event_KILL, events.HandlerFunc(func(e *executor.Event) error {
			log.Println("warning: KILL not implemented")
			return nil
		})),
		events.Handle(executor.Event_ACKNOWLEDGED, events.HandlerFunc(func(e *executor.Event) error {
			delete(state.unackedTasks, e.Acknowledged.TaskID)
			delete(state.unackedUpdates, string(e.Acknowledged.UUID))
			return nil
		})),
		events.Handle(executor.Event_MESSAGE, events.HandlerFunc(func(e *executor.Event) error {
			log.Printf("MESSAGE: received %d bytes of message data", len(e.Message.Data))
			return nil
		})),
		events.Handle(executor.Event_SHUTDOWN, events.HandlerFunc(func(e *executor.Event) error {
			log.Println("SHUTDOWN received")
			state.shouldQuit = true
			return nil
		})),
		events.Handle(executor.Event_ERROR, events.HandlerFunc(func(e *executor.Event) error {
			log.Println("ERROR received")
			return errMustAbort
		})),
	)
}

func sendFailedTasks(state *executorState) {
	for taskID, status := range state.failedTasks {
		updateErr := update(state, status)
		if updateErr != nil {
			log.Printf("failed to send status update for task %s: %+v", taskID.Value, updateErr)
		} else {
			delete(state.failedTasks, taskID)
		}
	}
}

func launch(state *executorState, task mesos.TaskInfo) {
	state.unackedTasks[task.TaskID] = task

	// send RUNNING
	status := newStatus(state, task.TaskID)
	status.State = mesos.TASK_RUNNING.Enum()
	err := update(state, status)
	if err != nil {
		log.Printf("failed to send TASK_RUNNING for task %s: %+v", task.TaskID.Value, err)
		status.State = mesos.TASK_FAILED.Enum()
		status.Message = protoString(err.Error())
		state.failedTasks[task.TaskID] = status
		return
	}

	// send FINISHED
	status = newStatus(state, task.TaskID)
	status.State = mesos.TASK_FINISHED.Enum()
	err = update(state, status)
	if err != nil {
		log.Printf("failed to send TASK_FINISHED for task %s: %+v", task.TaskID.Value, err)
		status.State = mesos.TASK_FAILED.Enum()
		status.Message = protoString(err.Error())
		state.failedTasks[task.TaskID] = status
	}
}

func protoString(s string) *string { return &s }

func update(state *executorState, status mesos.TaskStatus) error {
	upd := calls.Update(status).With(state.callOptions...)
	resp, err := state.cli.Do(upd)
	if resp != nil {
		resp.Close()
	}
	if err != nil {
		log.Printf("failed to send update: %+v", err)
	} else {
		state.unackedUpdates[string(status.UUID)] = *upd.Update
	}
	return err
}

func newStatus(state *executorState, id mesos.TaskID) mesos.TaskStatus {
	return mesos.TaskStatus{
		TaskID:     id,
		Source:     mesos.SOURCE_EXECUTOR.Enum(),
		ExecutorID: &state.executor.ExecutorID,
		UUID:       []byte(extras.Uuid()),
	}
}

type executorState struct {
	callOptions    executor.CallOptions
	cli            *httpcli.Client
	cfg            config.Config
	framework      mesos.FrameworkInfo
	executor       mesos.ExecutorInfo
	agent          mesos.AgentInfo
	unackedTasks   map[mesos.TaskID]mesos.TaskInfo
	unackedUpdates map[string]executor.Call_Update
	failedTasks    map[mesos.TaskID]mesos.TaskStatus
	shouldQuit     bool
}

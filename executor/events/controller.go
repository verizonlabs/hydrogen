package events

import (
	e "mesos-framework-sdk/executor"
	"mesos-framework-sdk/executor/events"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
	"time"
)

const (
	subscribeRetry = 2
)

type SprintExecutorController struct {
	executor  e.Executor
	logger    logging.Logger
	eventChan chan *exec.Event
}

func NewSprintExecutorEventController(e e.Executor, l logging.Logger) events.ExecutorEvents {
	return &SprintExecutorController{
		executor:  e,
		eventChan: make(chan *exec.Event),
		logger:    l,
	}
}

func (d *SprintExecutorController) Run() {
	go func() {
		for {
			err := d.executor.Subscribe(d.eventChan)
			if err != nil {
				d.logger.Emit(logging.ERROR, "Failed to subscribe: %s", err.Error())
				time.Sleep(time.Duration(subscribeRetry) * time.Second)
			}
		}
	}()

	select {
	case e := <-d.eventChan:
		d.Subscribed(e.GetSubscribed())
	}
	d.Listen()
}

// Default listening method on the
func (d *SprintExecutorController) Listen() {
	for {
		switch t := <-d.eventChan; t.GetType() {
		case exec.Event_SUBSCRIBED:
			d.Subscribed(t.GetSubscribed())
		case exec.Event_ACKNOWLEDGED:
			d.Acknowledged(t.GetAcknowledged())
		case exec.Event_MESSAGE:
			d.Message(t.GetMessage())
		case exec.Event_KILL:
			d.Kill(t.GetKill())
		case exec.Event_LAUNCH:
			d.Launch(t.GetLaunch())
		case exec.Event_LAUNCH_GROUP:
			d.LaunchGroup(t.GetLaunchGroup())
		case exec.Event_SHUTDOWN:
			d.Shutdown()
		case exec.Event_ERROR:
			d.Error(t.GetError())
		case exec.Event_UNKNOWN:
			d.logger.Emit(logging.INFO, "Unknown event caught")
		}
	}
}

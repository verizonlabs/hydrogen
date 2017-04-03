package events

/*
Handles events for the Executor
*/
import (
	"fmt"
	e "mesos-framework-sdk/executor"
	"mesos-framework-sdk/executor/events"
	exec "mesos-framework-sdk/include/executor"
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

func NewSprintExecutorEventController(
	exec e.Executor,
	eventChan chan *exec.Event,
	lgr logging.Logger) events.ExecutorEvents {

	return &SprintExecutorController{
		executor:  exec,
		eventChan: eventChan,
		logger:    lgr,
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
			d.logger.Emit(logging.ALARM, "Unknown event caught")
		}
	}
}

func (d *SprintExecutorController) Subscribed(sub *exec.Event_Subscribed) {
	sub.GetFrameworkInfo()
}

func (d *SprintExecutorController) Launch(launch *exec.Event_Launch) {
	fmt.Println(launch.GetTask())
}

func (d *SprintExecutorController) LaunchGroup(launchGroup *exec.Event_LaunchGroup) {
	fmt.Println(launchGroup.GetTaskGroup())
}

func (d *SprintExecutorController) Kill(kill *exec.Event_Kill) {
	fmt.Printf("%v, %v\n", kill.GetTaskId(), kill.GetKillPolicy())
}
func (d *SprintExecutorController) Acknowledged(acknowledge *exec.Event_Acknowledged) {
	fmt.Printf("%v\n", acknowledge.GetTaskId())
}
func (d *SprintExecutorController) Message(message *exec.Event_Message) {
	fmt.Printf("%v\n", message.GetData())
}
func (d *SprintExecutorController) Shutdown() {

}
func (d *SprintExecutorController) Error(error *exec.Event_Error) {
	fmt.Printf("%v\n", error.GetMessage())
}

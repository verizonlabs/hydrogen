package events

/*
Handles events for the Executor
*/
import (
	"fmt"
	exec "mesos-framework-sdk/include/executor"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"sprint/executor"
	"time"
)

const (
	subscribeRetry = 2
)

type SprintExecutorController struct {
	executor  *executor.SprintExecutor
	logger    logging.Logger
	eventChan chan *exec.Event
	kv        *etcd.Etcd
}

func NewSprintExecutorEventController(exec *executor.SprintExecutor, eventChan chan *exec.Event, lgr logging.Logger, kv *etcd.Etcd) *SprintExecutorController {
	return &SprintExecutorController{
		executor:  exec,
		eventChan: eventChan,
		logger:    lgr,
		kv:        kv,
	}

}

func (d *SprintExecutorController) Run() {
	id, err := d.kv.Read("/frameworkId")
	if err == nil {
		d.executor.FrameworkID = &mesos_v1.FrameworkID{Value: &id}
	}

	go func() {
		for {
			err = d.executor.Subscribe(d.eventChan)
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
			d.executor.FrameworkID = t.GetSubscribed().GetFrameworkInfo().GetId()
			go d.Subscribed(t.GetSubscribed())
		case exec.Event_ACKNOWLEDGED:
			go d.Acknowledged(t.GetAcknowledged())
		case exec.Event_MESSAGE:
			go d.Message(t.GetMessage())
		case exec.Event_KILL:
			go d.Kill(t.GetKill())
		case exec.Event_LAUNCH:
			go d.Launch(t.GetLaunch())
		case exec.Event_LAUNCH_GROUP:
			go d.LaunchGroup(t.GetLaunchGroup())
		case exec.Event_SHUTDOWN:
			go d.Shutdown()
		case exec.Event_ERROR:
			go d.Error(t.GetError())
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

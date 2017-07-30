package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	mockLogger "mesos-framework-sdk/logging/test"
	sdkResourceManager "mesos-framework-sdk/resources/manager"
	mockResourceManager "mesos-framework-sdk/resources/manager/test"
	sdkScheduler "mesos-framework-sdk/scheduler"
	sched "mesos-framework-sdk/scheduler/test"
	"os"
	"sprint/scheduler"
	sprintTask "sprint/task/manager"
	mockTaskManager "sprint/task/manager/test"
	"sprint/task/persistence"
	mockStorage "sprint/task/persistence/test"
	"testing"
	"time"
)

var (
	states []*mesos_v1.TaskState = []*mesos_v1.TaskState{
		mesos_v1.TaskState_TASK_RUNNING.Enum(),
		mesos_v1.TaskState_TASK_STARTING.Enum(),
		mesos_v1.TaskState_TASK_FINISHED.Enum(),
		mesos_v1.TaskState_TASK_DROPPED.Enum(),
		mesos_v1.TaskState_TASK_ERROR.Enum(),
		mesos_v1.TaskState_TASK_FAILED.Enum(),
		mesos_v1.TaskState_TASK_GONE.Enum(),
		mesos_v1.TaskState_TASK_GONE_BY_OPERATOR.Enum(),
		mesos_v1.TaskState_TASK_UNREACHABLE.Enum(),
		mesos_v1.TaskState_TASK_UNKNOWN.Enum(),
		mesos_v1.TaskState_TASK_STAGING.Enum(),
		mesos_v1.TaskState_TASK_KILLED.Enum(),
		mesos_v1.TaskState_TASK_KILLING.Enum(),
		mesos_v1.TaskState_TASK_ERROR.Enum(),
		mesos_v1.TaskState_TASK_LOST.Enum(),
	}
)

// TODO (tim): Factory function that takes in a list of broken items,
// generates an event controller with those items broken.

// Creates a new working event controller.
func workingEventController() *SprintEventController {
	var (
		cfg *scheduler.Configuration = &scheduler.Configuration{
			Leader: &scheduler.LeaderConfiguration{
				IP: "1", // Make sure we break out of our HA loop by matching on what mock storage gives us.
			},
			Executor:  &scheduler.ExecutorConfiguration{},
			Scheduler: &scheduler.SchedulerConfiguration{ReconcileInterval: time.Nanosecond},
			Persistence: &scheduler.PersistenceConfiguration{
				MaxRetries: 0,
			},
		}
		c  chan *mesos_v1_scheduler.Event     = make(chan *mesos_v1_scheduler.Event)
		sh sdkScheduler.Scheduler             = sched.MockScheduler{}
		m  sprintTask.SprintTaskManager       = &mockTaskManager.MockTaskManager{}
		rm sdkResourceManager.ResourceManager = &mockResourceManager.MockResourceManager{}
		s  persistence.Storage                = &mockStorage.MockStorage{}
		l  logging.Logger                     = &mockLogger.MockLogger{}
	)
	return &SprintEventController{
		config:          cfg,
		scheduler:       sh,
		taskmanager:     m,
		resourcemanager: rm,
		events:          c,
		storage:         s,
		logger:          l,
	}
}

func brokenSchedulerEventController() *SprintEventController {
	var (
		cfg *scheduler.Configuration = &scheduler.Configuration{
			Leader:    &scheduler.LeaderConfiguration{},
			Executor:  &scheduler.ExecutorConfiguration{},
			Scheduler: &scheduler.SchedulerConfiguration{ReconcileInterval: time.Nanosecond},
			Persistence: &scheduler.PersistenceConfiguration{
				MaxRetries: 0,
			},
		}
		c  chan *mesos_v1_scheduler.Event     = make(chan *mesos_v1_scheduler.Event)
		sh sdkScheduler.Scheduler             = sched.MockBrokenScheduler{}
		m  sprintTask.SprintTaskManager       = &mockTaskManager.MockTaskManager{}
		rm sdkResourceManager.ResourceManager = &mockResourceManager.MockResourceManager{}
		s  persistence.Storage                = &mockStorage.MockStorage{}
		l  logging.Logger                     = &mockLogger.MockLogger{}
	)
	return &SprintEventController{
		config:          cfg,
		scheduler:       sh,
		taskmanager:     m,
		resourcemanager: rm,
		events:          c,
		storage:         s,
		logger:          l,
	}
}

func TestNewSprintEventController(t *testing.T) {
	ctrl := workingEventController()
	if ctrl == nil {
		t.FailNow()
	}
}

func TestSprintEventController_Name(t *testing.T) {
	ctrl := workingEventController()
	_, err := ctrl.Name()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

/*
func TestSprintEventController_Run(t *testing.T) {
	ctrl := workingEventController()

	go ctrl.Run()

	// Subscribe.
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}
}

// This should utilize the broken factory
func TestSprintEventController_FailureToRun(t *testing.T) {
	ctrl := workingEventController()

	go ctrl.Run()

	// Subscribe.
	ctrl.events <- &mesos_v1_scheduler.Event{
		Type: mesos_v1_scheduler.Event_SUBSCRIBED.Enum(),
		Subscribed: &mesos_v1_scheduler.Event_Subscribed{
			FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("Test")},
		},
	}
}*/

func TestSprintEventController_SignalHandler(t *testing.T) {
	ctrl := workingEventController()
	ctrl.registerShutdownHandlers()
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(os.Interrupt)
	p.Wait()
}

func TestNewSprintEventController_periodicReconcile(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.periodicReconcile()
	broken := brokenSchedulerEventController()
	go broken.periodicReconcile()

}

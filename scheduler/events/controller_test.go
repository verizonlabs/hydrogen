package events

import (
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging/test"
	"mesos-framework-sdk/persistence/drivers/etcd/test"
	"mesos-framework-sdk/resources/manager/test"
	sched "mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/task/manager/test"
	"mesos-framework-sdk/utils"
	"sprint/scheduler"
	"testing"
)

func TestNewSprintEventController(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManager{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}

	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	if ctrl == nil {
		t.FailNow()
	}
}

func TestSprintEventController_Offers(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManager{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}

	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	// Test empty offers.
	offers := []*mesos_v1.Offer{}
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

	// Test a single offer
	offers = []*mesos_v1.Offer{}
	resources := []*mesos_v1.Resource{}
	resources = append(resources, &mesos_v1.Resource{
		Name: proto.String("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: proto.Float64(10.0),
		},
	})
	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: proto.String("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: proto.String(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: proto.String(utils.UuidAsString())},
		Hostname:    proto.String("Some host"),
		Resources:   resources,
	})
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

}

func TestSprintEventController_OffersWithQueuedTasks(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}

	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	// Test empty offers.
	offers := []*mesos_v1.Offer{}
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

	// Test a single offer
	offers = []*mesos_v1.Offer{}
	resources := []*mesos_v1.Resource{}
	resources = append(resources, &mesos_v1.Resource{
		Name: proto.String("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: proto.Float64(10.0),
		},
	})
	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: proto.String("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: proto.String(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: proto.String(utils.UuidAsString())},
		Hostname:    proto.String("Some host"),
		Resources:   resources,
	})
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

}

func TestSprintEventController_Name(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}

	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	_, err := ctrl.Name()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestSprintEventController_Update(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}

	states := []*mesos_v1.TaskState{
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
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: proto.String("id")},
			State:  state,
		}})
	}
}

func TestSprintEventController_Subscribe(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Subscribe(&mesos_v1_scheduler.Event_Subscribed{FrameworkId: &mesos_v1.FrameworkID{Value: proto.String("id")}})
}

func TestSprintEventController_CreateAndGetLeader(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.CreateLeader()
	ctrl.GetLeader()
}

/*
func TestSprintEventController_Run(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.NewMockScheduler()
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event, 10)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	t.Log("Testing Run...")
	ctrl.Run()
}
*/

func TestSprintEventController_TaskManager(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	a := ctrl.TaskManager()
	if a == nil {
		t.FailNow()
	}
}

func TestSprintEventController_Communicate(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	go ctrl.Communicate() // this never gets covered since this test closes too fast
	ctrl.Election()
}

func TestSprintEventController_Message(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: proto.String("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: proto.String("id")},
		Data:       []byte(`some message`),
	})
}

func TestSprintEventController_Failure(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Failure(&mesos_v1_scheduler.Event_Failure{
		AgentId: &mesos_v1.AgentID{Value: proto.String("agent")},
	})
}

func TestSprintEventController_InverseOffer(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.InverseOffer(&mesos_v1_scheduler.Event_InverseOffers{
		InverseOffers: []*mesos_v1.InverseOffer{
			{
				Id:             &mesos_v1.OfferID{Value: proto.String("id")},
				FrameworkId:    &mesos_v1.FrameworkID{Value: proto.String("id")},
				Unavailability: &mesos_v1.Unavailability{Start: &mesos_v1.TimeInfo{Nanoseconds: proto.Int64(0)}},
			},
		},
	})
}

func TestSprintEventController_RescindInverseOffer(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.RescindInverseOffer(&mesos_v1_scheduler.Event_RescindInverseOffer{
		InverseOfferId: &mesos_v1.OfferID{Value: proto.String("id")},
	})
}

func TestSprintEventController_Error(t *testing.T) {
	rm := MockResourceManager.MockResourceManager{}
	mg := testTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := test.MockEtcd{}
	lg := MockLogging.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Error(&mesos_v1_scheduler.Event_Error{
		Message: proto.String("message"),
	})
}

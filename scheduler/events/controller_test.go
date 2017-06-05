package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	mockLogger "mesos-framework-sdk/logging/test"
	mockResourceManager "mesos-framework-sdk/resources/manager/test"
	sched "mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/utils"
	"sprint/scheduler"
	mockTaskManager "sprint/task/manager/test"
	mockStorage "sprint/task/persistence/test"
	"testing"
)

func TestNewSprintEventController(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManager{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}

	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	if ctrl == nil {
		t.FailNow()
	}
}

func TestSprintEventController_Offers(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManager{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader:   &scheduler.LeaderConfiguration{},
		Executor: &scheduler.ExecutorConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}

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
		Name: utils.ProtoString("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: utils.ProtoFloat64(10.0),
		},
	})
	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: utils.ProtoString("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: utils.ProtoString(utils.UuidAsString())},
		Hostname:    utils.ProtoString("Some host"),
		Resources:   resources,
	})
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

}

func TestSprintEventController_OffersWithQueuedTasks(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader:   &scheduler.LeaderConfiguration{},
		Executor: &scheduler.ExecutorConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}

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
		Name: utils.ProtoString("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: utils.ProtoFloat64(10.0),
		},
	})
	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: utils.ProtoString("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: utils.ProtoString(utils.UuidAsString())},
		Hostname:    utils.ProtoString("Some host"),
		Resources:   resources,
	})
	ctrl.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

}

func TestSprintEventController_Name(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}

	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	_, err := ctrl.Name()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestSprintEventController_Update(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}

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
			TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
			State:  state,
		}})
	}
}

func TestSprintEventController_Subscribe(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Subscribed(&mesos_v1_scheduler.Event_Subscribed{FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString("id")}})
}

func TestSprintEventController_CreateAndGetLeader(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.CreateLeader()
	ctrl.GetLeader()
}

func TestSprintEventController_Communicate(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	go ctrl.Communicate() // this never gets covered since this test closes too fast
	ctrl.Election()
}

func TestSprintEventController_Message(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Message(&mesos_v1_scheduler.Event_Message{
		AgentId:    &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		ExecutorId: &mesos_v1.ExecutorID{Value: utils.ProtoString("id")},
		Data:       []byte(`some message`),
	})
}

func TestSprintEventController_Failure(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Failure(&mesos_v1_scheduler.Event_Failure{
		AgentId: &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
	})
}

func TestSprintEventController_InverseOffer(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.InverseOffer(&mesos_v1_scheduler.Event_InverseOffers{
		InverseOffers: []*mesos_v1.InverseOffer{
			{
				Id:             &mesos_v1.OfferID{Value: utils.ProtoString("id")},
				FrameworkId:    &mesos_v1.FrameworkID{Value: utils.ProtoString("id")},
				Unavailability: &mesos_v1.Unavailability{Start: &mesos_v1.TimeInfo{Nanoseconds: utils.ProtoInt64(0)}},
			},
		},
	})
}

func TestSprintEventController_RescindInverseOffer(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.RescindInverseOffer(&mesos_v1_scheduler.Event_RescindInverseOffer{
		InverseOfferId: &mesos_v1.OfferID{Value: utils.ProtoString("id")},
	})
}

func TestSprintEventController_Error(t *testing.T) {
	rm := mockResourceManager.MockResourceManager{}
	mg := mockTaskManager.MockTaskManagerQueued{}
	sh := sched.MockScheduler{}
	cfg := &scheduler.Configuration{
		Leader: &scheduler.LeaderConfiguration{},
	}
	eventChan := make(chan *mesos_v1_scheduler.Event)
	kv := mockStorage.MockStorage{}
	lg := mockLogger.MockLogger{}
	ctrl := NewSprintEventController(cfg, sh, mg, rm, eventChan, kv, lg)
	ctrl.Error(&mesos_v1_scheduler.Event_Error{
		Message: utils.ProtoString("message"),
	})
}

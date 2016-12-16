package scheduler

import (
	"mesos-sdk"
	sched "mesos-sdk/scheduler"
	"reflect"
	"testing"
)

// Prepare common data for our tests.
func init() {
	s = &mockScheduler{
		cfg: cfg,
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{
				Value: "test",
			},
		},
		state: state{
			frameworkId: "test",
		},
	}
	event = &sched.Event{
		Subscribed: &sched.Event_Subscribed{
			FrameworkID: &mesos.FrameworkID{
				Value: s.State().frameworkId,
			},
		},
		Failure: &sched.Event_Failure{
			ExecutorID: &s.ExecutorInfo().ExecutorID,
		},
	}
}

// Makes sure we get our handlers back correctly.
func TestNewHandlers(t *testing.T) {
	t.Parallel()

	h := NewHandlers(s)
	if reflect.TypeOf(h) != reflect.TypeOf(new(handlers)) {
		t.Fatal("Handlers is of the wrong type")
	}
	if err := h.mux.HandleEvent(event); err != nil {
		t.Fatal("Handler mux failed to handle event: " + err.Error())
	}
	if err := h.ack.HandleEvent(event); err != nil {
		t.Fatal("Handler ack failed to handle event: " + err.Error())
	}
}

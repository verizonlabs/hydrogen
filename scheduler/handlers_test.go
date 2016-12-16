package scheduler

import (
	"mesos-sdk"
	ev "mesos-sdk/scheduler/events"
	sched "mesos-sdk/scheduler"
	"reflect"
	"testing"
)

// Mocked handlers.
type mockHandlers struct {
	sched scheduler
	mux *ev.Mux
	ack ev.Handler
}

func (m *mockHandlers) Mux() *ev.Mux {
	return m.mux
}

func (m *mockHandlers) Ack() ev.Handler {
	return m.ack
}

func (m *mockHandlers) ResourceOffers(offers []mesos.Offer) error {
	return nil
}

var h handlers

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
	h = NewHandlers(s)
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

	if reflect.TypeOf(h) != reflect.TypeOf(new(sprintHandlers)) {
		t.Fatal("Handlers is of the wrong type")
	}
	if err := h.Mux().HandleEvent(event); err != nil {
		t.Fatal("Handler mux failed to handle event: " + err.Error())
	}
	if err := h.Ack().HandleEvent(event); err != nil {
		t.Fatal("Handler ack failed to handle event: " + err.Error())
	}
}

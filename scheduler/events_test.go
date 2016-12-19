package scheduler

import (
	"mesos-sdk"
	sched "mesos-sdk/scheduler"
	"mesos-sdk/scheduler/calls"
	ev "mesos-sdk/scheduler/events"
	"testing"
)

//Mocked events.
type mockEvents struct{}

func (m *mockEvents) Subscribed(event *sched.Event) error {
	return nil
}

func (m *mockEvents) Offers(event *sched.Event) error {
	return nil
}

func (m *mockEvents) Update(event *sched.Event) error {
	return nil
}

func (m *mockEvents) Failure(event *sched.Event) error {
	return nil
}

var (
	e     events
	event *sched.Event
)

// Prepare common data for our tests.
func init() {
	cfg = new(mockConfiguration).Initialize(nil)
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
	h = &mockHandlers{
		sched: s,
		ack: ev.AcknowledgeUpdates(func() calls.Caller {
			return *s.Caller()
		}),
	}
	e = NewEvents(s, h.Ack(), h)
	event = &sched.Event{
		Subscribed: &sched.Event_Subscribed{
			FrameworkID: &mesos.FrameworkID{
				Value: s.State().frameworkId,
			},
		},
		Failure: &sched.Event_Failure{
			ExecutorID: &s.ExecutorInfo().ExecutorID,
		},
		Offers: &sched.Event_Offers{
			Offers: []mesos.Offer{
				{
					FrameworkID: mesos.FrameworkID{
						Value: "test",
					},
					AgentID: mesos.AgentID{
						Value: "test",
					},
				},
			},
		},
	}
}

// Makes sure we get the correct type back for the events.
func TestNewEvents(t *testing.T) {
	t.Parallel()

	switch e.(type) {
	case *sprintEvents:
		return
	default:
		t.Fatal("Controller is not of the right type")
	}
}

// Measures performance for getting our events.
func BenchmarkNewEvents(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewEvents(s, h.Ack(), h)
	}
}

// Checks the subscribed event handler.
func TestSprintEvents_Subscribed(t *testing.T) {
	t.Parallel()

	if err := e.Subscribed(event); err != nil {
		t.Fatal("Subscribed event failure: " + err.Error())
	}
}

// Measures performance of the subscribed event.
func BenchmarkSprintEvents_Subscribed(b *testing.B) {
	for n := 0; n < b.N; n++ {
		e.Subscribed(event)
	}
}

// Checks the offers event handler.
func TestSprintEvents_Offers(t *testing.T) {
	t.Parallel()

	if err := e.Offers(event); err != nil {
		t.Fatal("Offers event failure: " + err.Error())
	}
}

// Measures performance of the offers event.
func BenchmarkSprintEvents_Offers(b *testing.B) {
	for n := 0; n < b.N; n++ {
		e.Offers(event)
	}
}

// Checks the update event handler.
func TestSprintEvents_Update(t *testing.T) {
	t.Parallel()

	if err := e.Update(event); err != nil {
		t.Fatal("Update event failure: " + err.Error())
	}
}

// Checks the failure event handler
func TestSprintEvents_Failure(t *testing.T) {
	t.Parallel()

	if err := e.Failure(event); err != nil {
		t.Fatal("Failure event failed: " + err.Error())
	}
}

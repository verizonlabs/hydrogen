package scheduler

import (
	"fmt"
	"mesos-sdk"
	sched "mesos-sdk/scheduler"
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

var e events

//Prepare common data for our tests.
func init() {
	cfg = new(mockConfiguration).Initialize(nil)
	s = &mockScheduler{
		cfg: cfg,
		state: state{
			frameworkId: "test",
		},
	}
	e = NewEvents(s, NewHandlers(s).ack)
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

func TestSprintEvents_Subscribed(t *testing.T) {
	t.Parallel()

	event := &sched.Event{
		Subscribed: &sched.Event_Subscribed{
			FrameworkID: &mesos.FrameworkID{
				Value: "test",
			},
		},
	}

	fmt.Println(event.GetSubscribed().GetFrameworkID().GetValue())

	if err := e.Subscribed(event); err != nil {
		t.Fatal("Subscribed event failure: " + err.Error())
	}
}

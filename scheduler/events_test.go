package scheduler

import (
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

package scheduler

import (
	"mesos-sdk"
	sched "mesos-sdk/scheduler"
	"reflect"
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
	e     events       = &mockEvents{}
	event *sched.Event = &sched.Event{
		Subscribed: &sched.Event_Subscribed{
			FrameworkID: &mesos.FrameworkID{
				Value: "test",
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
)

// Makes sure we get the correct type back for the events.
func TestNewEvents(t *testing.T) {
	t.Parallel()

	e := NewEvents(s, h.Ack(), h)
	if reflect.TypeOf(e) != reflect.TypeOf(new(sprintEvents)) {
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

	e := NewEvents(s, h.Ack(), h)

	if err := e.Subscribed(event); err != nil {
		t.Fatal("Subscribed event failure: " + err.Error())
	}
}

// Measures performance of the subscribed event.
func BenchmarkSprintEvents_Subscribed(b *testing.B) {
	e := NewEvents(s, h.Ack(), h)

	for n := 0; n < b.N; n++ {
		e.Subscribed(event)
	}
}

// Checks the offers event handler.
func TestSprintEvents_Offers(t *testing.T) {
	t.Parallel()

	e := NewEvents(s, h.Ack(), h)

	if err := e.Offers(event); err != nil {
		t.Fatal("Offers event failure: " + err.Error())
	}
}

// Measures performance of the offers event.
func BenchmarkSprintEvents_Offers(b *testing.B) {
	e := NewEvents(s, h.Ack(), h)

	for n := 0; n < b.N; n++ {
		e.Offers(event)
	}
}

// Checks the update event handler.
func TestSprintEvents_Update(t *testing.T) {
	t.Parallel()

	e := NewEvents(s, h.Ack(), h)

	if err := e.Update(event); err != nil {
		t.Fatal("Update event failure: " + err.Error())
	}
}

// Measures performance of the update event.
func BenchmarkSprintEvents_Update(b *testing.B) {
	e := NewEvents(s, h.Ack(), h)

	for n := 0; n < b.N; n++ {
		e.Update(event)
	}
}

// Checks the failure event handler
func TestSprintEvents_Failure(t *testing.T) {
	t.Parallel()

	e := NewEvents(s, h.Ack(), h)

	if err := e.Failure(event); err != nil {
		t.Fatal("Failure event failed: " + err.Error())
	}
}

// Measures performance of the failure event.
func BenchmarkSprintEvents_Failure(b *testing.B) {
	e := NewEvents(s, h.Ack(), h)

	for n := 0; n < b.N; n++ {
		e.Failure(event)
	}
}

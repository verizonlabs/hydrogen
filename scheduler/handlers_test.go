package scheduler

import (
	"mesos-sdk"
	"mesos-sdk/scheduler/calls"
	ev "mesos-sdk/scheduler/events"
	"reflect"
	"testing"
)

// Mocked handlers.
type mockHandlers struct {
	sched scheduler
	mux   *ev.Mux
	ack   ev.Handler
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

var h handlers = &mockHandlers{
	sched: s,
	ack: ev.AcknowledgeUpdates(func() calls.Caller {
		return *s.Caller()
	}),
}

// Makes sure we get our handlers back correctly.
func TestNewHandlers(t *testing.T) {
	t.Parallel()

	h := NewHandlers(s)

	if reflect.TypeOf(h) != reflect.TypeOf(new(sprintHandlers)) {
		t.Fatal("Handlers is of the wrong type")
	}
}

// Handler multiplexer tests.
func TestSprintHandlers_Mux(t *testing.T) {
	t.Parallel()

	h := NewHandlers(s)

	if reflect.TypeOf(h.Mux()) != reflect.TypeOf(new(ev.Mux)) {
		t.Fatal("Handler multiplexer has the wrong type")
	}
}

// Acknowledgement handler tests.
func TestSprintHandlers_Ack(t *testing.T) {
	t.Parallel()

	h := NewHandlers(s)

	if reflect.TypeOf(h.Ack()) != reflect.TypeOf(ev.AcknowledgeUpdates(nil)) {
		t.Fatal("Acknowledgement handler has the wrong type")
	}
}

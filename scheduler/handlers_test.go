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

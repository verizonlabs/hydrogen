package scheduler

import (
	ctrl "mesos-sdk/extras/scheduler/controller"
	sched "mesos-sdk/scheduler"
	"mesos-sdk/scheduler/calls"
	ev "mesos-sdk/scheduler/events"
)

// Holds context about our event multiplexer and acknowledge handler.
type handlers struct {
	mux *ev.Mux
	ack ev.Handler
}

// Sets up function handlers to process incoming events from Mesos.
func NewHandlers(s scheduler) *handlers {
	ack := ev.AcknowledgeUpdates(func() calls.Caller {
		return *s.Caller()
	})

	events := NewEvents(s, ack)

	return &handlers{
		mux: ev.NewMux(
			ev.DefaultHandler(ev.HandlerFunc(ctrl.DefaultHandler)),
			ev.MapFuncs(map[sched.Event_Type]ev.HandlerFunc{
				sched.Event_SUBSCRIBED: events.Subscribed,
				sched.Event_OFFERS:     events.Offers,
				sched.Event_UPDATE:     events.Update,
				sched.Event_FAILURE:    events.Failure,
			}),
		),
		ack: ack,
	}
}

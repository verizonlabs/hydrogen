package scheduler

import (
	ctrl "github.com/verizonlabs/mesos-go/extras/scheduler/controller"
	sched "github.com/verizonlabs/mesos-go/scheduler"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	ev "github.com/verizonlabs/mesos-go/scheduler/events"
)

// Holds context about our event multiplexer and acknowledge handler.
type handlers struct {
	mux *ev.Mux
	ack ev.Handler
}

// Sets up function handlers to process incoming events from Mesos.
func NewHandlers(s scheduler) *handlers {
	ack := ev.AcknowledgeUpdates(func() calls.Caller {
		return *s.GetCaller()
	})

	events := NewEvents(s, ack)

	return &handlers{
		mux: ev.NewMux(
			ev.DefaultHandler(ev.HandlerFunc(ctrl.DefaultHandler)),
			ev.MapFuncs(map[sched.Event_Type]ev.HandlerFunc{
				sched.Event_SUBSCRIBED: events.subscribed,
				sched.Event_OFFERS:     events.offers,
				sched.Event_UPDATE:     events.update,
				sched.Event_FAILURE:    events.failure,
			}),
		),
		ack: ack,
	}
}

package scheduler

import (
	"errors"
	"log"
	sched "mesos-sdk/scheduler"
	"mesos-sdk/scheduler/calls"
	ev "mesos-sdk/scheduler/events"
	"strconv"
)

// Base implementation for Mesos event handlers.
type events interface {
	Subscribed(event *sched.Event) error
	Offers(event *sched.Event) error
	Update(event *sched.Event) error
	Failure(event *sched.Event) error
}

// Holds context about our scheduler and acknowledge handler.
type sprintEvents struct {
	sched    scheduler
	ack      ev.Handler
	handlers handlers
}

// Applies the contextual information from the scheduler.
func NewEvents(s scheduler, a ev.Handler, h handlers) *sprintEvents {
	return &sprintEvents{
		sched:    s,
		ack:      a,
		handlers: h,
	}
}

// Handler for subscribed events.
func (e *sprintEvents) Subscribed(event *sched.Event) error {
	log.Println("Received subscribe event")
	if e.sched.State().frameworkId == "" {
		e.sched.State().frameworkId = event.GetSubscribed().GetFrameworkID().GetValue()
		if e.sched.State().frameworkId == "" {
			return errors.New("mesos gave us an empty frameworkID")
		} else {
			*e.sched.Caller() = calls.FrameworkCaller(e.sched.State().frameworkId).Apply(*e.sched.Caller())
		}
	}
	return nil
}

// Handler for offers events.
func (e *sprintEvents) Offers(event *sched.Event) error {
	offers := event.GetOffers().GetOffers()
	err := e.handlers.ResourceOffers(offers)
	if err != nil {
		log.Println("Error handling resource offers: " + err.Error())
	}

	return err
}

// Handler for update events.
func (e *sprintEvents) Update(event *sched.Event) error {
	log.Println("Received update event")
	if err := e.ack.HandleEvent(event); err != nil {
		log.Println("Failed to acknowledge status update for task: " + err.Error())
	}
	// TODO handle status updates
	return nil
}

// Handler for failure events.
func (e *sprintEvents) Failure(event *sched.Event) error {
	log.Println("Received failure event")
	f := event.GetFailure()
	if f.ExecutorID != nil {
		msg := "Executor '" + f.ExecutorID.Value + "' terminated"
		if f.AgentID != nil {
			msg += " on agent '" + f.AgentID.Value + "'"
		}
		if f.Status != nil {
			msg += " with status=" + strconv.Itoa(int(*f.Status))
		}
		log.Println(msg)
	} else if f.AgentID != nil {
		log.Println("Agent '" + f.AgentID.Value + "' terminated")
	}
	return nil
}

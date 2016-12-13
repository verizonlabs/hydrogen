package scheduler

import (
	"errors"
	sched "github.com/verizonlabs/mesos-go/scheduler"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	ev "github.com/verizonlabs/mesos-go/scheduler/events"
	"log"
	"strconv"
)

// Holds context about our scheduler and acknowledge handler.
type events struct {
	sched baseScheduler
	ack   ev.Handler
}

// Applies the contextual information from the scheduler.
func NewEvents(s baseScheduler, a ev.Handler) *events {
	return &events{
		sched: s,
		ack:   a,
	}
}

// Handler for subscribed events
func (e *events) subscribed(event *sched.Event) error {
	log.Println("Received subscribe event")
	if e.sched.GetState().frameworkId == "" {
		e.sched.GetState().frameworkId = event.GetSubscribed().GetFrameworkID().GetValue()
		if e.sched.GetState().frameworkId == "" {
			return errors.New("mesos gave us an empty frameworkID")
		} else {
			*e.sched.GetCaller() = calls.FrameworkCaller(e.sched.GetState().frameworkId).Apply(*e.sched.GetCaller())
		}
	}
	return nil
}

// Handler for offers events
func (e *events) offers(event *sched.Event) error {
	//TODO implement handling resource offers
	return nil
}

// Handler for update events
func (e *events) update(event *sched.Event) error {
	log.Println("Received update event")
	if err := e.ack.HandleEvent(event); err != nil {
		log.Println("Failed to acknowledge status update for task: " + err.Error())
	}
	// TODO handle status updates
	return nil
}

// Handler for failure events
func (e *events) failure(event *sched.Event) error {
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

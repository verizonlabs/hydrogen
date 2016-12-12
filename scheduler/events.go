package scheduler

import (
	"errors"
	sched "github.com/verizonlabs/mesos-go/scheduler"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	ev "github.com/verizonlabs/mesos-go/scheduler/events"
	"log"
	"strconv"
)

type events struct {
	sched *scheduler
	ack   ev.Handler
}

func NewEvents(s *scheduler, a ev.Handler) *events {
	return &events{
		sched: s,
		ack:   a,
	}
}

func (e *events) subscribed(event *sched.Event) error {
	log.Println("Received subscribe event")
	if e.sched.state.frameworkId == "" {
		e.sched.state.frameworkId = event.GetSubscribed().GetFrameworkID().GetValue()
		if e.sched.state.frameworkId == "" {
			return errors.New("mesos gave us an empty frameworkID")
		} else {
			e.sched.http = calls.FrameworkCaller(e.sched.state.frameworkId).Apply(e.sched.http)
		}
	}
	return nil
}

func (e *events) offers(event *sched.Event) error {
	//TODO implement handling resource offers
	return nil
}

func (e *events) update(event *sched.Event) error {
	log.Println("Received update event")
	if err := e.ack.HandleEvent(event); err != nil {
		log.Println("Failed to acknowledge status update for task: " + err.Error())
	}
	// TODO handle status updates
	return nil
}

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

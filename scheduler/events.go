package scheduler

/*
Events handle all events that come through to the scheduler from the mesos-master.
*/

import (
	"fmt"
	"log"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/include/scheduler"
	"strconv"
)

type SprintEvents struct {
	FrameworkId *mesos_v1.FrameworkID
}

// Applies the contextual information from the scheduler.
func NewEvents() *SprintEvents {
	return &SprintEvents{}
}

// Handler for subscribed events.
func (e *SprintEvents) Subscribe(subEvent *mesos_v1_scheduler.Event_Subscribed) {
	log.Println("Received subscribe event")

	if subEvent.GetFrameworkId().GetValue() == "" {
		e.FrameworkId = subEvent.GetFrameworkId()
		if e.FrameworkId.GetValue() == "" {
			log.Println("mesos gave us an empty frameworkID")
		} else {
			log.Println("Scheduler's framework ID is " + e.FrameworkId.GetValue())
		}
	}
}

func (e *SprintEvents) Rescind(*mesos_v1_scheduler.Event_Rescind) {

}

// Offer event
func (e *SprintEvents) Offers(eventOffers *mesos_v1_scheduler.Event_Offers) {
	offers := eventOffers.GetOffers()
	for k, v := range offers {
		fmt.Printf("%v: %v", k, v)
	}
	//err := e.handlers.ResourceOffers(offers)
}

// Handler for update events.
func (e *SprintEvents) Update(updateEvent *mesos_v1_scheduler.Event_Update) {
	log.Println("Received update event")
	fmt.Printf("%v", updateEvent)
	//e.handlers.StatusUpdates(updateEvent.GetStatus())
}

// Handler for failure events.
func (e *SprintEvents) Failure(failureEvent *mesos_v1_scheduler.Event_Failure) {
	log.Println("Received failure event")

	if failureEvent.GetExecutorId().GetValue() != "" {
		msg := "Executor '" + failureEvent.GetExecutorId().GetValue() + "' terminated"
		if failureEvent.GetAgentId().GetValue() != "" {
			msg += " on agent '" + failureEvent.GetAgentId().GetValue() + "'"
		}
		if failureEvent.GetStatus() != 0 {
			msg += " with status=" + strconv.Itoa(int(failureEvent.GetStatus()))
		}
		log.Println(msg)
	} else if failureEvent.GetAgentId().GetValue() != "" {
		log.Println("Agent '" + failureEvent.GetAgentId().GetValue() + "' terminated")
	}
}

func (e *SprintEvents) InverseOffer(*mesos_v1_scheduler.Event_InverseOffers) {

}
func (e *SprintEvents) RescindInverseOffer(*mesos_v1_scheduler.Event_RescindInverseOffer) {

}
func (e *SprintEvents) Message(*mesos_v1_scheduler.Event_Message) {

}
func (e *SprintEvents) Error(*mesos_v1_scheduler.Event_Error) {

}

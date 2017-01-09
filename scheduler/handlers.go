package scheduler

import (
	"math/rand"
	"mesos-sdk"
	ctrl "mesos-sdk/extras/scheduler/controller"
	sched "mesos-sdk/scheduler"
	"mesos-sdk/scheduler/calls"
	ev "mesos-sdk/scheduler/events"
	"strconv"
	"time"
	"log"
)

type handlers interface {
	Mux() *ev.Mux
	Ack() ev.Handler
	ResourceOffers(offers []mesos.Offer) error
	StatusUpdates(mesos.TaskStatus)
}

// Holds context about our event multiplexer and acknowledge handler.
type sprintHandlers struct {
	sched scheduler
	mux   *ev.Mux
	ack   ev.Handler
}

// Sets up function handlers to process incoming events from Mesos.
func NewHandlers(s scheduler) *sprintHandlers {
	ack := ev.AcknowledgeUpdates(func() calls.Caller {
		return *s.Caller()
	})

	handlers := &sprintHandlers{}
	events := NewEvents(s, ack, handlers)
	handlers.sched = s
	handlers.mux = ev.NewMux(
		ev.DefaultHandler(ev.HandlerFunc(ctrl.DefaultHandler)),
		ev.MapFuncs(map[sched.Event_Type]ev.HandlerFunc{
			sched.Event_SUBSCRIBED: events.Subscribed,
			sched.Event_OFFERS:     events.Offers,
			sched.Event_UPDATE:     events.Update,
			sched.Event_FAILURE:    events.Failure,
		}),
	)
	handlers.ack = ack

	return handlers
}

// Returns the handler's multiplexer.
func (h *sprintHandlers) Mux() *ev.Mux {
	return h.mux
}

// Returns the handler's acknowledgement handler.
func (h *sprintHandlers) Ack() ev.Handler {
	return h.ack
}

// Handler for our received resource offers.
func (h *sprintHandlers) ResourceOffers(offers []mesos.Offer) error {
	log.Println("In resource offers...")
	jitter := rand.New(rand.NewSource(time.Now().Unix()))
	callOption := calls.RefuseSecondsWithJitter(jitter, h.sched.Config().MaxRefuse())
	state := h.sched.State()

	for i := range offers {
		var (
			remaining = mesos.Resources(offers[i].Resources)
			tasks     = []mesos.TaskInfo{}
		)

		log.Println("received offer id '" + offers[i].ID.Value + "' with resources " + remaining.String())

		var executorResources mesos.Resources
		if len(offers[i].ExecutorIDs) == 0 {
			executorResources = mesos.Resources(h.sched.ExecutorInfo().Resources)
		}
		log.Println("Executor resources: " + executorResources.String())

		flattened := remaining.Flatten()
		taskResources := state.taskResources.Plus(executorResources...)

		log.Println("Tasks launched: " + strconv.Itoa(state.tasksLaunched))
		log.Println("RESOURCES: " + taskResources.String())
		log.Println("RESOURCES FLAT: " + flattened.String())
		log.Println("Total tasks: " + strconv.Itoa(state.totalTasks))

		for state.tasksLaunched < state.totalTasks && flattened.ContainsAll(taskResources) {
			state.tasksLaunched++
			taskId := state.tasksLaunched

			task := mesos.TaskInfo{
				TaskID: mesos.TaskID{
					Value: strconv.Itoa(taskId),
				},
				AgentID:   offers[i].AgentID,
				Executor:  h.sched.ExecutorInfo(),
				Resources: remaining.Find(taskResources.Flatten(mesos.Role(state.role).Assign())),
			}
			task.Name = "task_" + task.TaskID.Value

			remaining.Subtract(task.Resources...)
			tasks = append(tasks, task)

			flattened = remaining.Flatten()
		}

		log.Println("Generating an accept call")
		log.Println(tasks)
		accept := calls.Accept(
			calls.OfferOperations{
				calls.OpLaunch(tasks...),
			}.WithOffers(offers[i].ID),
		).With(callOption)

		log.Println("POST accept call")
		log.Println(accept.String())
		err := calls.CallNoData(*h.sched.Caller(), accept)
		if err != nil {
			log.Printf("failed to launch tasks: %+v", err)
		}
	}
	return nil
}

// Handler for status updates from Mesos.
func (h *sprintHandlers) StatusUpdates(s mesos.TaskStatus) {
	state := h.sched.State()

	switch st := s.GetState(); st {
	case mesos.TASK_FINISHED:
		state.tasksFinished++

		if state.tasksFinished == state.totalTasks {
			h.sched.SuppressOffers()
		} else {
			h.sched.ReviveOffers()
		}
	case mesos.TASK_LOST:
		// TODO Handle task lost.
	case mesos.TASK_KILLED:
		// TODO Handle task killed.
	case mesos.TASK_FAILED:
		// TODO Handle task failed.
	case mesos.TASK_ERROR:
		// TODO Handle task error.
	}
}

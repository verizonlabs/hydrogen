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
)

// Holds context about our event multiplexer and acknowledge handler.
type handlers struct {
	sched scheduler
	mux   *ev.Mux
	ack   ev.Handler
}

// Sets up function handlers to process incoming events from Mesos.
func NewHandlers(s scheduler) *handlers {
	ack := ev.AcknowledgeUpdates(func() calls.Caller {
		return *s.Caller()
	})

	events := NewEvents(s, ack)

	handlers := &handlers{
		sched: s,
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

	events.setHandlers(handlers)
	return handlers
}

// Handler for our received resource offers.
func (h *handlers) resourceOffers(offers []mesos.Offer) error {
	jitter := rand.New(rand.NewSource(time.Now().Unix()))
	callOption := calls.RefuseSecondsWithJitter(jitter, h.sched.Config().MaxRefuse())
	state := h.sched.State()

	for i := range offers {
		var (
			remaining = mesos.Resources(offers[i].Resources)
			tasks     = []mesos.TaskInfo{}
		)

		var executorResources mesos.Resources
		if len(offers[i].ExecutorIDs) == 0 {
			executorResources = mesos.Resources(h.sched.ExecutorInfo().Resources)
		}

		flattened := remaining.Flatten()
		taskResources := state.taskResources.Plus(executorResources...)

		for state.tasksLaunched < state.totalTasks && flattened.ContainsAll(taskResources) {
			state.tasksLaunched++
			taskId := state.tasksLaunched

			task := mesos.TaskInfo{
				TaskID: mesos.TaskID{
					Value: strconv.Itoa(taskId),
				},
				AgentID:   offers[i].AgentID,
				Executor:  h.sched.ExecutorInfo(),
				Resources: remaining.Find(state.taskResources.Flatten(mesos.Role(state.role).Assign())),
			}
			task.Name = "task_" + task.TaskID.Value

			remaining.Subtract(task.Resources...)
			tasks = append(tasks, task)

			flattened = remaining.Flatten()
		}

		accept := calls.Accept(
			calls.OfferOperations{
				calls.OpLaunch(tasks...),
			}.WithOffers(offers[i].ID),
		).With(callOption)

		err := calls.CallNoData(*h.sched.Caller(), accept)
		if err != nil {
			return err
		}
	}
	return nil
}

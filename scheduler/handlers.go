package scheduler

import (
	"math/rand"
	"mesos-sdk"
	"mesos-sdk/extras"
	ctrl "mesos-sdk/extras/controller"
	sched "mesos-sdk/scheduler"
	"mesos-sdk/scheduler/calls"
	ev "mesos-sdk/scheduler/events"
	"time"
)

type handlers interface {
	Mux() *ev.Mux
	Ack() ev.Handler
	ResourceOffers(offers []mesos.Offer) error
	StatusUpdates(mesos.TaskStatus)
}

// Holds context about our event multiplexer and acknowledge handler.
type sprintHandlers struct {
	sched Scheduler
	mux   *ev.Mux
	ack   ev.Handler
}

// Sets up function handlers to process incoming events from Mesos.
func NewHandlers(s Scheduler) *sprintHandlers {
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
		// TODO build support for specifying resources
		// This will be eventually hooked up to the API where users can submit tasks
		// In this way we can parameterize the executor/task resources (per task too)
		// We also want to combine the task resources with the executor resources
		taskResources := state.taskResources.Plus(executorResources...)

		// TODO don't use the length of the map once we have our API
		// Detect the number of jobs the user wants to launch instead
		for len(state.tasks) < state.totalTasks && flattened.ContainsAll(taskResources) {
			// Ignore the error here since we know that we're generating a valid v4 UUID.
			// Other people using their own UUIDs should probably check this.
			uuid, _ := extras.UuidToString(extras.Uuid())

			task := mesos.TaskInfo{
				TaskID: mesos.TaskID{
					// TaskID needs to be unique.
					Value: uuid,
				},
				AgentID:  offers[i].AgentID,
				Executor: h.sched.NewExecutor(),
				// TODO once the resource parameterization is in place reference state.taskResources again
				// Right now state.taskResources is empty which will cause issues
				Resources: remaining.Find(taskResources.Flatten(mesos.Role(state.role).Assign())),
			}
			task.Name = "sprinter_" + task.TaskID.Value

			remaining.Subtract(task.Resources...)

			tasks = append(tasks, task)
			state.tasks[task.TaskID.Value] = task.AgentID.Value

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

// Handler for status updates from Mesos.
func (h *sprintHandlers) StatusUpdates(s mesos.TaskStatus) {
	state := h.sched.State()

	switch st := s.GetState(); st {
	case mesos.TASK_FINISHED:
		// TODO we may not need this anymore once the API is complete
		state.tasksFinished++
		delete(state.tasks, s.GetTaskID().Value)

		// TODO change and move this check elsewhere once we have the API hooked up
		// We won't have to rely on these counters anymore and can just revive when a user submits a job
		// Suppression can be done in another spot when we determine that all tasks are launched
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

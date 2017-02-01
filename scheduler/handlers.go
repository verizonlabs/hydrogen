package scheduler

import (
	"math/rand"
	"mesos-sdk"
	ctrl "mesos-sdk/extras/scheduler/controller"
	sched "mesos-sdk/scheduler"
	"mesos-sdk/scheduler/calls"
	ev "mesos-sdk/scheduler/events"
	"mesos-sdk/taskmngr"
	"stash.verizon.com/dkt/mlog"
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
	sched SprintScheduler
	mux   *ev.Mux
	ack   ev.Handler
}

// Sets up function handlers to process incoming events from Mesos.
func NewHandlers(s SprintScheduler) *sprintHandlers {
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
	manager := h.sched.TaskManager()

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

		if ok, _ := manager.HasQueuedTasked(); ok {
			for task := range manager.Tasks().IterBuffered() {
				if flattened.ContainsAll(taskResources) {
					t, err := task.Val.(taskmngr.Task).Info()
					if err != nil {
						mlog.Error(err.Error())
					}

					taskInfo := t.(mesos.TaskInfo)
					taskInfo.AgentID = offers[i].AgentID
					taskInfo.Executor = h.sched.NewExecutor()
					taskInfo.Name = "sprint_" + taskInfo.TaskID.Value

					// Add it to the list of tasks to send off.
					tasks = append(tasks, taskInfo)
					remaining.Subtract(taskInfo.Resources...)
					flattened = remaining.Flatten()
				} else {
					break // No resources left, break out of the loop.
				}
			}

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
	switch st := s.GetState(); st {
	case mesos.TASK_FINISHED:
		if anyleft, _ := h.sched.taskmgr.HasQueuedTasked(); !anyleft {
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

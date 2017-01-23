package scheduler

import (
	"io"
	"log"
	"mesos-sdk"
	"mesos-sdk/backoff"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/scheduler/calls"
	"time"
)

// Sane defaults for backoff timings
var (
	RegistrationMinBackoff = 1 * time.Second
	RegistrationMaxBackoff = 15 * time.Second
)

// Base implementation of a controller
type controller interface {
	SchedulerCtrl() ctrl.Controller
	Scheduler() Scheduler
	BuildContext() *ctrl.ContextAdapter
	BuildFrameworkInfo(cfg configuration) *mesos.FrameworkInfo
	BuildConfig(ctx *ctrl.ContextAdapter, http *calls.Caller, shutdown <-chan struct{}, h handlers) *ctrl.Config
}

// Manages the context and configuration for our scheduler.
type sprintController struct {
	scheduler     Scheduler
	schedulerCtrl ctrl.Controller
	context       *ctrl.ContextAdapter
	config        *ctrl.Config
	shutdown      <-chan struct{}
}

// Returns a new controller with some shared state applied from the scheduler.
func NewController(s Scheduler, shutdown <-chan struct{}) *sprintController {
	return &sprintController{
		scheduler:     s,
		schedulerCtrl: ctrl.New(),
		shutdown:      shutdown,
	}
}

// Returns the internal scheduler controller.
func (c *sprintController) SchedulerCtrl() ctrl.Controller {
	return c.schedulerCtrl
}

// Returns the internal scheduler
func (c *sprintController) Scheduler() Scheduler {
	return c.scheduler
}

// Builds out context for us to use when managing state in the scheduler.
func (c *sprintController) BuildContext() *ctrl.ContextAdapter {
	c.context = &ctrl.ContextAdapter{
		DoneFunc: func() bool {
			return c.scheduler.State().done
		},
		FrameworkIDFunc: func() string {
			return c.scheduler.State().frameworkId
		},
		ErrorFunc: func(err error) {
			if err != nil && err != io.EOF {
				log.Println(err)
			} else {
				log.Println("Disconnected")
			}
		},
	}
	return c.context
}

// Builds out information about our framework that will be sent to Mesos.
func (c *sprintController) BuildFrameworkInfo(cfg configuration) *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{
		Name:       cfg.Name(),
		Checkpoint: cfg.Checkpointing(),
	}
}

// Builds out the controller configuration which uses our context and framework information.
func (c *sprintController) BuildConfig(ctx *ctrl.ContextAdapter, http *calls.Caller, shutdown <-chan struct{}, h handlers) *ctrl.Config {
	c.config = &ctrl.Config{
		Context:            ctx,
		Framework:          c.scheduler.FrameworkInfo(),
		Caller:             *http,
		RegistrationTokens: backoff.Notifier(RegistrationMinBackoff, RegistrationMaxBackoff, shutdown),
		Handler:            h.Mux(),
	}
	return c.config
}

package scheduler

import (
	"github.com/verizonlabs/mesos-go"
	"github.com/verizonlabs/mesos-go/backoff"
	ctrl "github.com/verizonlabs/mesos-go/extras/scheduler/controller"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	"io"
	"log"
	"time"
)

// Sane defaults for backoff timings
var (
	RegistrationMinBackoff = 1 * time.Second
	RegistrationMaxBackoff = 15 * time.Second
)

// Base implementation of a controller
type baseController interface {
	GetSchedulerCtrl() ctrl.Controller
	BuildContext() *ctrl.ContextAdapter
	BuildFrameworkInfo(cfg baseConfiguration) *mesos.FrameworkInfo
	BuildConfig(ctx *ctrl.ContextAdapter, cfg *mesos.FrameworkInfo, http *calls.Caller, shutdown <-chan struct{}, h *handlers) *ctrl.Config
}

// Manages the context and configuration for our scheduler.
type controller struct {
	scheduler     baseScheduler
	schedulerCtrl ctrl.Controller
	context       *ctrl.ContextAdapter
	config        *ctrl.Config
	shutdown      <-chan struct{}
}

// Returns a new controller with some shared state applied from the scheduler.
func NewController(s baseScheduler, shutdown <-chan struct{}) *controller {
	return &controller{
		scheduler:     s,
		schedulerCtrl: ctrl.New(),
		shutdown:      shutdown,
	}
}

// Returns the internal scheduler controller.
func (c *controller) GetSchedulerCtrl() ctrl.Controller {
	return c.schedulerCtrl
}

// Builds out context for us to use when managing state in the scheduler.
func (c *controller) BuildContext() *ctrl.ContextAdapter {
	c.context = &ctrl.ContextAdapter{
		DoneFunc: func() bool {
			return c.scheduler.GetState().done
		},
		FrameworkIDFunc: func() string {
			return c.scheduler.GetState().frameworkId
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
func (c *controller) BuildFrameworkInfo(cfg baseConfiguration) *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{
		Name:       cfg.GetName(),
		Checkpoint: cfg.GetCheckpointing(),
	}
}

// Builds out the controller configuration which uses our context and framework information.
func (c *controller) BuildConfig(ctx *ctrl.ContextAdapter, cfg *mesos.FrameworkInfo, http *calls.Caller, shutdown <-chan struct{}, h *handlers) *ctrl.Config {
	c.config = &ctrl.Config{
		Context:            ctx,
		Framework:          cfg,
		Caller:             *http,
		RegistrationTokens: backoff.Notifier(RegistrationMinBackoff, RegistrationMaxBackoff, shutdown),
		Handler:            h.mux,
	}
	return c.config
}

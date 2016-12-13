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

var (
	RegistrationMinBackoff = 1 * time.Second
	RegistrationMaxBackoff = 15 * time.Second
)

// Manages the context and configuration for our scheduler.
type controller struct {
	scheduler     *scheduler
	schedulerCtrl ctrl.Controller
	context       *ctrl.ContextAdapter
	config        *ctrl.Config
	shutdown      <-chan struct{}
}

// Returns a new controller with some shared state applied from the scheduler.
func NewController(s *scheduler, shutdown <-chan struct{}) *controller {
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
			return c.scheduler.state.done
		},
		FrameworkIDFunc: func() string {
			return c.scheduler.state.frameworkId
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
func (c *controller) BuildFrameworkInfo(cfg *Configuration) *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{
		Name:       cfg.name,
		Checkpoint: &cfg.checkpointing,
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

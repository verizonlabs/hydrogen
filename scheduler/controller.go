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

type controller struct {
	scheduler     *scheduler
	schedulerCtrl ctrl.Controller
	context       *ctrl.ContextAdapter
	config        *ctrl.Config
	shutdown      <-chan struct{}
}

func NewController(s *scheduler, shutdown <-chan struct{}) *controller {
	return &controller{
		scheduler:     s,
		schedulerCtrl: ctrl.New(),
		shutdown:      shutdown,
	}
}

func (c *controller) GetSchedulerCtrl() ctrl.Controller {
	return c.schedulerCtrl
}

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

func (c *controller) BuildFrameworkInfo(cfg *Configuration) *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{
		Name:       cfg.name,
		Checkpoint: &cfg.checkpointing,
	}
}

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

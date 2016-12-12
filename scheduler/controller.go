package scheduler

import (
	"github.com/verizonlabs/mesos-go"
	"github.com/verizonlabs/mesos-go/backoff"
	ctrl "github.com/verizonlabs/mesos-go/extras/scheduler/controller"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	"io"
	"time"
)

var (
	RegistrationMinBackoff = 1 * time.Second
	RegistrationMaxBackoff = 15 * time.Second
)

type controller struct {
	schedulerCtrl ctrl.Controller
	context       *ctrl.ContextAdapter
	config        ctrl.Config
	shutdown      <-chan struct{}
}

func NewController(shutdown <-chan struct{}) *controller {
	return &controller{
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
			//TODO implement state for this
			return false
		},
		FrameworkIDFunc: func() string {
			//TODO implement state for this
			return ""
		},
		ErrorFunc: func(err error) {
			if err != nil && err != io.EOF {
				//TODO determine how we want to log across the project
			} else {
				//TODO determine how we want to log across the project
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

func (c *controller) BuildConfig(ctx *ctrl.ContextAdapter, cfg *mesos.FrameworkInfo, http *calls.Caller, shutdown <-chan struct{}) ctrl.Config {
	c.config = ctrl.Config{
		Context:            ctx,
		Framework:          cfg,
		Caller:             *http,
		RegistrationTokens: backoff.Notifier(RegistrationMinBackoff, RegistrationMaxBackoff, shutdown),
		//Handler:            handler,
	}
	return c.config
}

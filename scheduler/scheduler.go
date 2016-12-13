package scheduler

import (
	"github.com/gogo/protobuf/proto"
	"github.com/verizonlabs/mesos-go"
	"github.com/verizonlabs/mesos-go/backoff"
	"github.com/verizonlabs/mesos-go/encoding"
	ctrl "github.com/verizonlabs/mesos-go/extras/scheduler/controller"
	"github.com/verizonlabs/mesos-go/httpcli"
	"github.com/verizonlabs/mesos-go/httpcli/httpsched"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	"net/http"
	"time"
)

// Base implementation of a scheduler.
type baseScheduler interface {
	Run(c ctrl.Controller, config *ctrl.Config) error
	GetState() *state
	GetCaller() *calls.Caller
}

// Scheduler state.
type state struct {
	frameworkId   string
	tasksLaunched uint
	tasksFinished uint
	totalTasks    uint
	done          bool
	reviveTokens  <-chan struct{}
}

// Holds all necessary information for our scheduler to function.
type scheduler struct {
	config    *Configuration
	framework *mesos.FrameworkInfo
	executor  *mesos.ExecutorInfo
	http      calls.Caller
	shutdown  chan struct{}
	state     state
}

// Returns a new scheduler using user-supplied configuration.
func NewScheduler(cfg *Configuration, shutdown chan struct{}) *scheduler {
	return &scheduler{
		config: cfg,
		framework: &mesos.FrameworkInfo{
			Name:       cfg.name,
			Checkpoint: &cfg.checkpointing,
		},
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{
				Value: "default",
			},
			Name: proto.String("Sprinter"),
			Command: mesos.CommandInfo{
				Value: proto.String(cfg.command),
				URIs:  cfg.uris,
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_MESOS.Enum(),
			},
		},
		http: httpsched.NewCaller(httpcli.New(
			httpcli.Endpoint(cfg.endpoint),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.timeout),
					httpcli.Transport(func(t *http.Transport) {
						t.ResponseHeaderTimeout = 15 * time.Second
						t.MaxIdleConnsPerHost = 2
					}),
				),
			),
		)),
		shutdown: shutdown,
		state: state{
			reviveTokens: backoff.BurstNotifier(cfg.reviveBurst, cfg.reviveWait, cfg.reviveWait, nil),
		},
	}
}

// Returns the internal state of the scheduler
func (s *scheduler) GetState() *state {
	return &s.state
}

// Returns the caller that we use for communication.
func (s *scheduler) GetCaller() *calls.Caller {
	return &s.http
}

// Runs our scheduler with some applied configuration.
func (s *scheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return c.Run(*config)
}

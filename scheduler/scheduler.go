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
type scheduler interface {
	Run(c ctrl.Controller, config *ctrl.Config) error
	GetState() *state
	GetCaller() *calls.Caller
	GetFrameworkInfo() *mesos.FrameworkInfo
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
type sprintScheduler struct {
	config    configuration
	framework *mesos.FrameworkInfo
	executor  *mesos.ExecutorInfo
	http      calls.Caller
	shutdown  chan struct{}
	state     state
}

// Returns a new scheduler using user-supplied configuration.
func NewScheduler(cfg configuration, shutdown chan struct{}) *sprintScheduler {
	return &sprintScheduler{
		config: cfg,
		framework: &mesos.FrameworkInfo{
			Name:       cfg.GetName(),
			Checkpoint: cfg.GetCheckpointing(),
		},
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{
				Value: "default",
			},
			Name: proto.String("Sprinter"),
			Command: mesos.CommandInfo{
				Value: proto.String(cfg.GetCommand()),
				URIs:  cfg.GetUris(),
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_MESOS.Enum(),
			},
		},
		http: httpsched.NewCaller(httpcli.New(
			httpcli.Endpoint(cfg.GetEndpoint()),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.GetTimeout()),
					httpcli.Transport(func(t *http.Transport) {
						t.ResponseHeaderTimeout = 15 * time.Second
						t.MaxIdleConnsPerHost = 2
					}),
				),
			),
		)),
		shutdown: shutdown,
		state: state{
			reviveTokens: backoff.BurstNotifier(cfg.GetReviveBurst(), cfg.GetReviveWait(), cfg.GetReviveWait(), nil),
		},
	}
}

// Returns the internal state of the scheduler
func (s *sprintScheduler) GetState() *state {
	return &s.state
}

// Returns the caller that we use for communication.
func (s *sprintScheduler) GetCaller() *calls.Caller {
	return &s.http
}

// Returns the FrameworkInfo that is sent to Mesos.
func (s *sprintScheduler) GetFrameworkInfo() *mesos.FrameworkInfo {
	return s.framework
}

// Runs our scheduler with some applied configuration.
func (s *sprintScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return c.Run(*config)
}

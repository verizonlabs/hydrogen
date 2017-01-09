package scheduler

import (
	"mesos-sdk"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/scheduler/calls"
	"net/http"
	"time"
	"mesos-sdk/httpcli/httpsched"
	"github.com/gogo/protobuf/proto"
	"fmt"
)

// Base implementation of a scheduler.
type scheduler interface {
	Config() configuration
	Run(c ctrl.Controller, config *ctrl.Config) error
	State() *state
	Caller() *calls.Caller
	FrameworkInfo() *mesos.FrameworkInfo
	ExecutorInfo() *mesos.ExecutorInfo
	SuppressOffers() error
	ReviveOffers() error
}

// Scheduler state.
type state struct {
	frameworkId   string
	tasksLaunched int
	tasksFinished int
	totalTasks    int
	taskResources mesos.Resources
	role          string
	done          bool
	reviveTokens  <-chan struct{}
}

// Holds all nece  ssary information for our scheduler to function.
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
	var executorName = new(string)
	*executorName = "Sprinter"

	var isExecutable = new(bool)
	*isExecutable = true

	var uris = make([]mesos.CommandInfo_URI, 1)

	uris[0] = mesos.CommandInfo_URI{
		Value: "http://localhost:8081/executor",
		Executable: isExecutable,
	}

	var mesosContainer = new(mesos.ContainerInfo_Type)
	*mesosContainer = mesos.ContainerInfo_MESOS

	fmt.Print("Executor is at " + uris[0].Value + "\n")

	return &sprintScheduler{
		config: cfg,
		framework: &mesos.FrameworkInfo{
			Name:       cfg.Name(),
			Checkpoint: cfg.Checkpointing(),
		},
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{Value: "Default"},
			Name: proto.String("Test Executor"),
			Command: mesos.CommandInfo{
				Value: proto.String("echo hello"),
				Shell: proto.Bool(true),
				URIs: uris,
			},
			Resources: []mesos.Resource{
				{
					Name: "cpus",
					Type: mesos.SCALAR.Enum(),
					Scalar: &mesos.Value_Scalar{Value: 0.5},
				},
				{
					Name: "mem",
					Type: mesos.SCALAR.Enum(),
					Scalar: &mesos.Value_Scalar{Value: 1024.0},
				},
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_MESOS.Enum(),
				Mesos: &mesos.ContainerInfo_MesosInfo{
					Image: &mesos.Image{
						Docker: &mesos.Image_Docker{
							Name: "busybox:latest",
						},
						Type: mesos.Image_DOCKER.Enum(),
					},
				},

			},
		},
		http: httpsched.NewCaller(httpcli.New(
			httpcli.Endpoint(cfg.Endpoint()),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.Timeout()),
					httpcli.Transport(func(t *http.Transport) {
						t.ResponseHeaderTimeout = 15 * time.Second
						t.MaxIdleConnsPerHost = 2
					}),
				),
			),
		)),
		shutdown: shutdown,
		state: state{
			totalTasks: 5, // For testing, we need to allow POST'ing of tasks to the framework.
			reviveTokens: backoff.BurstNotifier(cfg.ReviveBurst(), cfg.ReviveWait(), cfg.ReviveWait(), nil),
		},
	}
}

// Returns the scheduler's configuration.
func (s *sprintScheduler) Config() configuration {
	return s.config
}

// Returns the internal state of the scheduler
func (s *sprintScheduler) State() *state {
	return &s.state
}

// Returns the caller that we use for communication.
func (s *sprintScheduler) Caller() *calls.Caller {
	return &s.http
}

// Returns the FrameworkInfo that is sent to Mesos.
func (s *sprintScheduler) FrameworkInfo() *mesos.FrameworkInfo {
	return s.framework
}

// Returns the ExecutorInfo that's associated with the scheduler.
func (s *sprintScheduler) ExecutorInfo() *mesos.ExecutorInfo {
	return s.executor
}

// Runs our scheduler with some applied configuration.
func (s *sprintScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return c.Run(*config)
}

// This call suppresses our offers received from Mesos.
func (s *sprintScheduler) SuppressOffers() error {
	return calls.CallNoData(s.http, calls.Suppress())
}

// This call revives our offers received from Mesos.
func (s *sprintScheduler) ReviveOffers() error {
	select {
	// Rate limit offer revivals.
	case <-s.state.reviveTokens:
		return calls.CallNoData(s.http, calls.Revive())
	default:
		return nil
	}
}

package scheduler

import (
	"mesos-sdk"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	"mesos-sdk/extras"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	"mesos-sdk/scheduler/calls"
	"net/http"
	"sprint"
	"strconv"
	"time"
)

// Base implementation of a scheduler.
type scheduler interface {
	NewExecutor() *mesos.ExecutorInfo
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
	// Ignore the error here since we know that we're generating a valid v4 UUID.
	// Other people using their own UUIDs should probably check this.
	uuid, _ := extras.UuidToString(extras.Uuid())
	// TODO don't hardcode IP
	executorUri := cfg.ExecutorSrvCfg().ExecutorSrvProtocol() + "://10.0.2.2:" + strconv.Itoa(cfg.ExecutorSrvCfg().ExecutorSrvPort()) + "/executor"

	return &sprintScheduler{
		config: cfg,
		framework: &mesos.FrameworkInfo{
			User:       cfg.User(),
			Name:       cfg.Name(),
			Checkpoint: cfg.Checkpointing(),
		},
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{Value: uuid},
			Name:       cfg.ExecutorName(),
			Command: mesos.CommandInfo{
				Value: cfg.ExecutorCmd(),
				URIs: []mesos.CommandInfo_URI{
					{
						Value:      executorUri,
						Executable: sprint.ProtoBool(true),
					},
				},
			},
			// TODO parameterize this
			Resources: []mesos.Resource{
				sprint.Resource("cpus", 0.5),
				sprint.Resource("mem", 1024.0),
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_MESOS.Enum(),
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
			totalTasks:   5, // TODO For testing, we need to allow POST'ing of tasks to the framework.
			reviveTokens: backoff.BurstNotifier(cfg.ReviveBurst(), cfg.ReviveWait(), cfg.ReviveWait(), nil),
		},
	}
}

// Returns a new ExecutorInfo with a new UUID.
// Reuses existing data from the scheduler's original ExecutorInfo.
func (s *sprintScheduler) NewExecutor() *mesos.ExecutorInfo {
	// Ignore the error here since we know that we're generating a valid v4 UUID.
	// Other people using their own UUIDs should probably check this.
	uuid, _ := extras.UuidToString(extras.Uuid())

	return &mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{Value: uuid},
		Name:       s.executor.Name,
		Command:    s.executor.Command,
		Resources:  s.executor.Resources,
		Container:  s.executor.Container,
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

package scheduler

import (
	"errors"
	"mesos-sdk"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	"mesos-sdk/extras"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	"mesos-sdk/scheduler/calls"
	"net/http"
	"strconv"
	"time"
)

// Base implementation of a scheduler.
type Scheduler interface {
	NewExecutor() *mesos.ExecutorInfo
	Config() configuration
	Run(c ctrl.Controller, config *ctrl.Config) error
	State() *state
	Caller() *calls.Caller
	FrameworkInfo() *mesos.FrameworkInfo
	ExecutorInfo() *mesos.ExecutorInfo
	SuppressOffers() error
	ReviveOffers() error
	Reconcile() (mesos.Response, error)
}

// Scheduler state.
type state struct {
	frameworkId string
	// TODO this field will need to be reworked once our API is in place
	// We might not need this since we are targeting long running jobs
	// The API could pass data for how many jobs a user has sent and then we can compare that with a running count
	tasksFinished int
	// TODO think about changing the reconcile call in the SDK
	// Currently a map[string]string is expected
	// We may want to use this elsewhere in our scheduler which might require more data
	// If that's the case the reconcile call should be changed to something like map[string]mesos.TaskInfo
	// If more data is needed the key can also be changed to whatever type we need to pull from
	tasks map[string]string
	// TODO remove this once we've hooked up the API to accept jobs
	totalTasks    int
	taskResources mesos.Resources
	role          string
	done          bool
	reviveTokens  <-chan struct{}
}

// Holds all necessary information for our scheduler to function.
type SprintScheduler struct {
	config    configuration
	framework *mesos.FrameworkInfo
	executor  *mesos.ExecutorInfo
	http      calls.Caller
	shutdown  chan struct{}
	state     state
}

// Returns all tasks known to the scheduler
func (s *state) Tasks() map[string]string {
	return s.tasks
}

// Returns the taskID's if any match.
func (s *state) TaskSearch(id string) (string, error) {
	var taskID string
	if taskID, ok := s.tasks[id]; ok {
		return taskID, nil
	}
	return taskID, errors.New("No taskid found for: " + id)
}

// Returns a new scheduler using user-supplied configuration.
func NewScheduler(cfg configuration, shutdown chan struct{}) *SprintScheduler {
	// Ignore the error here since we know that we're generating a valid v4 UUID.
	// Other people using their own UUIDs should probably check this.
	uuid, _ := extras.UuidToString(extras.Uuid())
	// TODO don't hardcode IP
	executorUri := cfg.ExecutorSrvCfg().Protocol() + "://10.0.2.2:" + strconv.Itoa(*cfg.ExecutorSrvCfg().Port()) + "/executor"

	return &SprintScheduler{
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
						Executable: extras.ProtoBool(true),
					},
				},
			},
			// TODO parameterize this
			Resources: []mesos.Resource{
				extras.Resource("cpus", 0.1),
				extras.Resource("mem", 128.0),
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
			tasks:        make(map[string]string),
			totalTasks:   5, // TODO For testing, we need to allow POST'ing of tasks to the framework.
			reviveTokens: backoff.BurstNotifier(cfg.ReviveBurst(), cfg.ReviveWait(), cfg.ReviveWait(), nil),
		},
	}
}

// Returns a new ExecutorInfo with a new UUID.
// Reuses existing data from the scheduler's original ExecutorInfo.
func (s *SprintScheduler) NewExecutor() *mesos.ExecutorInfo {
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
func (s *SprintScheduler) Config() configuration {
	return s.config
}

// Returns the internal state of the scheduler
func (s *SprintScheduler) State() *state {
	return &s.state
}

// Returns the caller that we use for communication.
func (s *SprintScheduler) Caller() *calls.Caller {
	return &s.http
}

// Returns the FrameworkInfo that is sent to Mesos.
func (s *SprintScheduler) FrameworkInfo() *mesos.FrameworkInfo {
	return s.framework
}

// Returns the ExecutorInfo that's associated with the scheduler.
func (s *SprintScheduler) ExecutorInfo() *mesos.ExecutorInfo {
	return s.executor
}

// Runs our scheduler with some applied configuration.
func (s *SprintScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return c.Run(*config)
}

// This call suppresses our offers received from Mesos.
func (s *SprintScheduler) SuppressOffers() error {
	return calls.CallNoData(s.http, calls.Suppress())
}

// This call will perform reconciliation for our tasks.
func (s *SprintScheduler) Reconcile() (mesos.Response, error) {
	return calls.Caller(s.http).Call(
		calls.Reconcile(
			calls.ReconcileTasks(s.state.tasks),
		),
	)
}

// This call revives our offers received from Mesos.
func (s *SprintScheduler) ReviveOffers() error {
	select {
	// Rate limit offer revivals.
	case <-s.state.reviveTokens:
		return calls.CallNoData(s.http, calls.Revive())
	default:
		return nil
	}
}

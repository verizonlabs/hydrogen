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
	"mesos-sdk/taskmngr"
	"net/http"
	"sprint/scheduler/taskmanager"
	"stash.verizon.com/dkt/mlog"
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
	TaskManager() *taskmanager.Manager
}

// Scheduler state.
type state struct {
	frameworkId   string
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
	taskmgr   *taskmanager.Manager
	http      calls.Caller
	shutdown  chan struct{}
	state     state
}

// Returns a new scheduler using user-supplied configuration.
func NewScheduler(cfg configuration, shutdown chan struct{}) *SprintScheduler {
	// Ignore the error here since we know that we're generating a valid v4 UUID.
	// Other people using their own UUIDs should probably check this.
	uuid, _ := extras.UuidToString(extras.Uuid())
	mlog.Info(cfg.ExecutorSrvCfg().GetEndpoint() + "/executor")

	scheduler := SprintScheduler{
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
						Value:      cfg.ExecutorSrvCfg().GetEndpoint() + "/executor",
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
		taskmgr:  taskmanager.NewManager(),
		state: state{
			reviveTokens: backoff.BurstNotifier(cfg.ReviveBurst(), cfg.ReviveWait(), cfg.ReviveWait(), nil),
		},
	}

	// Create the number of executors we want.
	for i := 0; i < cfg.Executors(); i++ {
		// Create new tasks and add them into the manager.
		uuid, _ = extras.UuidToString(extras.Uuid())
		name := "sprinter_" + string(i)

		task := mesos.TaskInfo{
			TaskID: mesos.TaskID{
				Value: uuid,
			},
			Executor: &mesos.ExecutorInfo{
				ExecutorID: mesos.ExecutorID{Value: uuid},
				Name:       extras.ProtoString(name),
				Command: mesos.CommandInfo{
					Value: cfg.ExecutorCmd(),
					URIs: []mesos.CommandInfo_URI{
						{
							Value:      cfg.ExecutorSrvCfg().GetEndpoint() + "/executor",
							Executable: extras.ProtoBool(true),
						},
					},
				},
				Resources: []mesos.Resource{
					extras.Resource("cpus", 0.1),
					extras.Resource("mem", 128.0),
				},
				Container: &mesos.ContainerInfo{
					Type: mesos.ContainerInfo_MESOS.Enum(),
				},
			},
			Resources: []mesos.Resource{
				extras.Resource("cpus", 0.1),
				extras.Resource("mem", 128.0),
			},
		}

		task.Name = "sprinter_" + task.TaskID.Value

		newTask := taskmanager.TaskState{}
		newTask.SetId(task.TaskID.Value)
		newTask.SetInfo(task)
		newTask.SetStatus(taskmngr.QUEUED)
		a, _ := newTask.Info()
		scheduler.TaskManager().Add(&newTask)
	}

	return &scheduler

}

// Returns a new ExecutorInfo with a new UUID.
// Reuses existing data from the scheduler's original ExecutorInfo.
func (s *SprintScheduler) NewExecutor() *mesos.ExecutorInfo {
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

// Returns the scheduler's configuration.
func (s *SprintScheduler) TaskManager() *taskmanager.Manager {
	return s.taskmgr
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
	tasks, err := taskmanager.TaskIdAgentIdMap(s.taskmgr.Tasks())
	if err != nil {
		mlog.Error(err.Error())
	}
	return calls.Caller(s.http).Call(
		calls.Reconcile(
			calls.ReconcileTasks(tasks),
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

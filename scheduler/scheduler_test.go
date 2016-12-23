package scheduler

import (
	"io/ioutil"
	"log"
	"mesos-sdk"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	"mesos-sdk/scheduler/calls"
	"os"
	"reflect"
	"testing"
	"time"
)

// Mocked scheduler.
type mockScheduler struct {
	cfg      configuration
	executor *mesos.ExecutorInfo
	state    state
}

func (m *mockScheduler) Config() configuration {
	return m.cfg
}

func (m *mockScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return nil
}

func (m *mockScheduler) State() *state {
	return &m.state
}

func (m *mockScheduler) Caller() *calls.Caller {
	s := httpsched.NewCaller(httpcli.New())
	return &s
}

func (m *mockScheduler) ExecutorInfo() *mesos.ExecutorInfo {
	return &mesos.ExecutorInfo{}
}

func (m *mockScheduler) FrameworkInfo() *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{}
}

var s scheduler = &mockScheduler{
	cfg: cfg,
	executor: &mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{
			Value: "",
		},
	},
}

// Suppress our logging and start the tests.
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)
	os.Exit(m.Run())
}

// Ensures we get the correct type back for the scheduler.
func TestNewScheduler(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	if reflect.TypeOf(s) != reflect.TypeOf(new(sprintScheduler)) {
		t.Fatal("Controller is not of the right type")
	}
}

// Checks the configuration stored inside of the scheduler.
func TestSprintScheduler_Config(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	cfg := s.Config()
	if reflect.TypeOf(cfg) != reflect.TypeOf(new(SprintConfiguration)) {
		t.Fatal("Scheduler configuration is of the wrong type")
	}
	if !*cfg.Checkpointing() {
		t.Fatal("Scheduler checkpointing is not set correctly")
	}
	if *cfg.Command() != "" {
		t.Fatal("Scheduler command is not set properly")
	}
	if cfg.Endpoint() != "http://127.0.0.1:5050/api/v1/scheduler" {
		t.Fatal("Scheduler endpoint is not set correctly")
	}
	if cfg.Name() != "Sprint" {
		t.Fatal("Scheduler name is not set correctly")
	}
	if cfg.MaxRefuse() != 5*time.Second {
		t.Fatal("Scheduler refusal time is not set correctly")
	}
	if cfg.ReviveBurst() != 3 {
		t.Fatal("Scheduler revive burst amount is not set correctly")
	}
	if cfg.ReviveWait() != 1*time.Second {
		t.Fatal("Scheduler revive wait period is not set correctly")
	}
	if cfg.Principal() != "Sprint" {
		t.Fatal("Scheduler principal is not set correctly")
	}
	if cfg.Timeout() != 20*time.Second {
		t.Fatal("Scheduler timeout is not set correctly")
	}
}

// Ensures the scheduler's state and contained information is correct.
func TestSprintScheduler_State(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	st := s.State()
	if reflect.TypeOf(st) != reflect.TypeOf(new(state)) {
		t.Fatal("Scheduler state is of the wrong type")
	}
	if st.tasksFinished != 0 || st.tasksLaunched != 0 || st.totalTasks != 0 {
		t.Fatal("Starting state of scheduler tasks is not correct")
	}
	if st.done {
		t.Fatal("Scheduler should not be marked as done here")
	}
	if st.frameworkId != "" || st.role != "" {
		t.Fatal("Framework ID and/or role has the wrong starting value")
	}

	token := <-st.reviveTokens
	if token != *new(struct{}) {
		t.Fatal("Couldn't drain revive tokens")
	}

	resource := mesos.Resource{
		Scalar: &mesos.Value_Scalar{
			Value: 64,
		},
		Disk: &mesos.Resource_DiskInfo{
			Volume: &mesos.Volume{
				ContainerPath: "/tmp",
			},
			Source: &mesos.Resource_DiskInfo_Source{
				Mount: &mesos.Resource_DiskInfo_Source_Mount{
					Root: "/tmp",
				},
			},
		},
	}

	st.taskResources.Add(resource)
	if err := st.taskResources.Validate(); err != nil {
		t.Fatal("Failed to validate task resources")
	}
}

// Tests to see if the scheduler has the right caller.
func TestSprintScheduler_Caller(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	caller := s.Caller()
	if reflect.TypeOf(*caller) != reflect.TypeOf(httpsched.NewCaller(httpcli.New())) {
		t.Fatal("Scheduler does not have the right kind of caller")
	}
}

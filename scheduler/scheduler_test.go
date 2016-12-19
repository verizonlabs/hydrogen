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
	state: state{
		frameworkId: "test",
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

// Ensures the scheduler's state and contained information is correct.
func TestScheduler_GetState(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	st := s.State()
	if reflect.TypeOf(st) != reflect.TypeOf(new(state)) {
		t.Fatal("Scheduler state is of the wrong type")
	}

	if st.tasksFinished != 0 || st.tasksLaunched != 0 || st.totalTasks != 0 {
		t.Fatal("Starting state of scheduler tasks is not correct")
	}
}

// Tests to see if the scheduler has the right caller.
func TestScheduler_GetCaller(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	caller := s.Caller()
	if reflect.TypeOf(*caller) != reflect.TypeOf(httpsched.NewCaller(httpcli.New())) {
		t.Fatal("Scheduler does not have the right kind of caller")
	}
}

package scheduler

import (
	"mesos-sdk"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	"mesos-sdk/scheduler/calls"
	"reflect"
	"testing"
)

// Mocked scheduler.
type mockScheduler struct {
	cfg configuration
}

func (m *mockScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return nil
}

func (m *mockScheduler) State() *state {
	return new(state)
}

func (m *mockScheduler) Caller() *calls.Caller {
	s := httpsched.NewCaller(httpcli.New())
	return &s
}

func (m *mockScheduler) FrameworkInfo() *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{}
}

var s scheduler

func init() {
	cfg = new(mockConfiguration).Initialize(nil)
	s = NewScheduler(cfg, make(chan struct{}))
}

// Ensures we get the correct type back for the scheduler.
func TestNewScheduler(t *testing.T) {
	t.Parallel()

	switch s.(type) {
	case *sprintScheduler:
		return
	default:
		t.Fatal("Controller is not of the right type")
	}
}

// Ensures the scheduler's state and contained information is correct.
func TestScheduler_GetState(t *testing.T) {
	t.Parallel()

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

	caller := s.Caller()
	if reflect.TypeOf(*caller) != reflect.TypeOf(httpsched.NewCaller(httpcli.New())) {
		t.Fatal("Scheduler does not have the right kind of caller")
	}
}

package scheduler

import (
	ctrl "github.com/verizonlabs/mesos-go/extras/scheduler/controller"
	"github.com/verizonlabs/mesos-go/httpcli"
	"github.com/verizonlabs/mesos-go/httpcli/httpsched"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	"reflect"
	"testing"
)

// Mocked scheduler.
type mockScheduler struct{}

func (m *mockScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return nil
}

func (m *mockScheduler) GetState() *state {
	return new(state)
}

func (m *mockScheduler) GetCaller() *calls.Caller {
	s := httpsched.NewCaller(httpcli.New())
	return &s
}

var s baseScheduler

func init() {
	cfg = new(mockConfiguration)
	cfg.Initialize(nil)
	s = NewScheduler(cfg, make(chan struct{}))
}

// Ensures we get the correct type back for the scheduler.
func TestNewScheduler(t *testing.T) {
	t.Parallel()

	switch s.(type) {
	case *scheduler:
		return
	default:
		t.Fatal("Controller is not of the right type")
	}
}

// Ensures the scheduler's state and contained information is correct.
func TestScheduler_GetState(t *testing.T) {
	t.Parallel()

	st := s.GetState()
	if reflect.TypeOf(st) != reflect.TypeOf(new(state)) {
		t.Fatal("Scheduler state is of the wrong type")
	}

	if st.tasksFinished != 0 || st.tasksLaunched != 0 || st.totalTasks != 0 {
		t.Fatal("Starting state of scheduler tasks is not correct")
	}
}

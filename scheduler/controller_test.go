package scheduler

import (
	ctrl "github.com/verizonlabs/mesos-go/extras/scheduler/controller"
	"github.com/verizonlabs/mesos-go/httpcli"
	"github.com/verizonlabs/mesos-go/httpcli/httpsched"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	"testing"
)

// Mocked scheduler
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

func TestNewController(t *testing.T) {
	t.Parallel()

	var c baseController

	c = NewController(new(mockScheduler), make(<-chan struct{}))
	switch c.(type) {
	case *controller:
		return
	default:
		t.Fatal("Controller is not of the right type")
	}
}

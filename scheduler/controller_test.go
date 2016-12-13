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

// Ensures that we get the correct type from creating a new controller.
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

// Ensures that we get the correct type from getting the internal scheduler controller.
func TestController_GetSchedulerCtrl(t *testing.T) {
	t.Parallel()

	var c baseController

	c = NewController(new(mockScheduler), make(<-chan struct{}))
	switch c.GetSchedulerCtrl().(type) {
	case ctrl.Controller:
		return
	default:
		t.Fatal("Scheduler controller is not of the right type")
	}
}

// Ensures we have the right types after building the context.
func TestController_BuildContext(t *testing.T) {
	t.Parallel()

	var c baseController

	c = NewController(new(mockScheduler), make(<-chan struct{}))
	ctx := c.BuildContext()

	if reflect.TypeOf(ctx) != reflect.TypeOf(new(ctrl.ContextAdapter)) {
		t.Fatal("Controller context is not of the right type")
	}
	if reflect.TypeOf(ctx.DoneFunc).Kind() != reflect.Func {
		t.Fatal("Context does not have a valid done function")
	}
	if reflect.TypeOf(ctx.FrameworkIDFunc).Kind() != reflect.Func {
		t.Fatal("Context does not have a valid FrameworkID function")
	}
	if reflect.TypeOf(ctx.ErrorFunc).Kind() != reflect.Func {
		t.Fatal("Context does not have a valid error function")
	}

	if reflect.TypeOf(ctx.FrameworkIDFunc()).Kind() != reflect.String {
		t.Fatal("FrameworkID function does not return the correct type")
	}
	if reflect.TypeOf(ctx.DoneFunc()).Kind() != reflect.Bool {
		t.Fatal("FrameworkID function does not return the correct type")
	}

}

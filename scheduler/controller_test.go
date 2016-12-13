package scheduler

import (
	"flag"
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

// Mocked configuration
type mockConfiguration struct {
	name          string
	checkpointing *bool
}

func (m *mockConfiguration) Initialize(fs *flag.FlagSet) {
	checkpointing := true

	m.name = "Test"
	m.checkpointing = &checkpointing
}

func (m *mockConfiguration) GetName() string {
	return m.name
}

func (m *mockConfiguration) GetCheckpointing() *bool {
	return m.checkpointing
}

var (
	c   baseController
	cfg baseConfiguration
)

func init() {
	c = NewController(new(mockScheduler), make(<-chan struct{}))
	cfg = new(mockConfiguration)
	cfg.Initialize(nil)
}

// Ensures that we get the correct type from creating a new controller.
func TestNewController(t *testing.T) {
	t.Parallel()

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

// Ensures that we have correctly build the FrameworkInfo that will be sent to Mesos.
func TestController_BuildFrameworkInfo(t *testing.T) {
	t.Parallel()

	info := c.BuildFrameworkInfo(cfg)
	if info.GetName() != "Test" {
		t.Fatal("FrameworkInfo has the wrong name")
	}
	if info.GetCheckpoint() != true {
		t.Fatal("FrameworkInfo does not have checkpointing set correctly")
	}
}
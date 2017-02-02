package executor

import (
	"io/ioutil"
	"log"
	"mesos-sdk"
	"mesos-sdk/extras"
	"os"
	"reflect"
	"testing"
)

// Mocked executor.
type mockExecutor struct {
	config   configuration
	executor mesos.ExecutorInfo
}

func (m *mockExecutor) Run() {}

var e executor = &mockExecutor{
	executor: mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{Value: "Mock executor"},
		Name:       extras.ProtoString("Mocker"),
		Resources: []mesos.Resource{
			extras.Resource("cpus", 0.5),
			extras.Resource("mem", 1024.0),
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_MESOS.Enum(),
		},
	},
}

// ENTRY POINT FOR ALL TESTS IN THIS PACKAGE
// Suppress our logging and start the tests.
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)
	os.Exit(m.Run())
}

// Make sure we initialize properly.
func TestNewExecutor(t *testing.T) {
	exec := NewExecutor(cfg)

	if reflect.TypeOf(exec) != reflect.TypeOf(new(sprintExecutor)) {
		t.Fatal("Executor is not of the right type")
	}
}

// Measures performance of creating an executor.
func BenchmarkNewExecutor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewExecutor(cfg)
	}
}

// Make sure we can actually run our executor.
func TestSprintExecutor_Run(t *testing.T) {
	exec := NewExecutor(cfg)
	exec.Run()
}

// Measure performance of running our executor.
func BenchmarkSprintExecutor_Run(b *testing.B) {
	exec := NewExecutor(cfg)
	for n := 0; n < b.N; n++ {
		exec.Run()
	}
}

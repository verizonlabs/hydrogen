package executor

import (
	"io/ioutil"
	"log"
	"mesos-sdk"
	"os"
	"sprint"
	"testing"
)

type mockExecutor struct {
	config   configuration
	executor mesos.ExecutorInfo
}

func (m *mockExecutor) Run() {}

var e executor = &mockExecutor{
	executor: mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{Value: "Mock executor"},
		Name:       sprint.ProtoString("Mocker"),
		Resources: []mesos.Resource{
			sprint.Resource("cpus", 0.5),
			sprint.Resource("mem", 1024.0),
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

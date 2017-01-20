package executor

import (
	"io/ioutil"
	"log"
	"mesos-sdk"
	"mesos-sdk/extras"
	"os"
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

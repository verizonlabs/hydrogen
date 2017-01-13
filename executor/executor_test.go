package executor

import (
	"io/ioutil"
	"log"
	"mesos-sdk"
	"mesos-sdk/executor/config"
	"os"
	"sprint"
	"strconv"
	"testing"
)

type mockExec struct {
	execInfo mesos.ExecutorInfo
	cfg      config.Config
}

var exec = mockExec{
	execInfo: mesos.ExecutorInfo{
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

/*
	Test our executor behavior here
*/
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)
	os.Exit(m.Run())
}

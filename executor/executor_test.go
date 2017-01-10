package executor

import (
	"io/ioutil"
	"log"
	"mesos-sdk"
	"mesos-sdk/executor/config"
	"os"
	"testing"
)

type mockExec struct {
	execInfo mesos.ExecutorInfo
	cfg      config.Config
}

var exec = mockExec{
	execInfo: mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{Value: "Mock executor"},
		Name:       ProtoString("Sprinter"),
		Command:    CommandInfo("echo 'hello world'"),
		Resources: []mesos.Resource{
			CpuResources(0.5),
			MemResources(1024.0),
		},
		Container: Container("busybox:latest"),
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

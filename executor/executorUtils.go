package executor

import (
	"github.com/verizonlabs/mesos-go/mesosproto"
)

func NewExecInfo() (*mesosproto.ExecutorInfo){
	cmd := new(string)
	*cmd = "/bin/echo 'Testing'"
	execid := new(string)
	frameworkId := new(string)
	*execid = "0123456"
	*frameworkId = "FrameworkID"

	return &mesosproto.ExecutorInfo{
		ExecutorId: &mesosproto.ExecutorID{Value: execid},
		FrameworkId: &mesosproto.FrameworkID{Value: frameworkId},
		Command: &mesosproto.CommandInfo{Value: cmd},
	}
}
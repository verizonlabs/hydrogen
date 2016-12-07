package executor

import (
	mesos "github.com/verizonlabs/mesos-go/mesosproto"
	exec "github.com/verizonlabs/mesos-go/executor"
)

type executor struct {
	launchedTasks int
}

func NewExecutor() *executor {
	return &executor{
		launchedTasks: 0,
	}
}

func (e *executor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwInfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	//Registered
}

func (e *executor) Reregistered(driver exec.ExecutorDriver, slaveInfo *mesos.SlaveInfo) {
	//Re-registered
}

func (e *executor) Disconnected(driver exec.ExecutorDriver) {
	//Executor disconnected
}

func (e *executor) LaunchTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	//Launch tasks
}

func (e *executor) KillTask(driver exec.ExecutorDriver, taskId *mesos.TaskID) {
	//Kill task
}

func (e *executor) FrameworkMessage(driver exec.ExecutorDriver, msg string) {
	//Messages
}

func (e *executor) Shutdown(driver exec.ExecutorDriver) {
	//Shutdown executor
}

func (e *executor) Error(driver exec.ExecutorDriver, err string) {
	//Executor got an error
}


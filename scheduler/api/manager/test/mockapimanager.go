package test

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/task/retry"
)

type (
	MockApiManager       struct{}
	MockBrokenApiManager struct{}
)

func (m MockApiManager) Deploy([]byte) ([]*manager.Task, error) {
	return []*manager.Task{
		{Info: &mesos_v1.TaskInfo{}},
	}, nil
}
func (m MockApiManager) Kill([]byte) (string, error) { return "", nil }
func (m MockApiManager) Update([]byte) ([]*manager.Task, error) {
	return []*manager.Task{{Info: &mesos_v1.TaskInfo{}}}, nil
}

func (m MockApiManager) Status(string) (mesos_v1.TaskState, error) {
	return mesos_v1.TaskState_TASK_RUNNING, nil
}
func (m MockApiManager) AllTasks() ([]manager.Task, error) {
	return []manager.Task{{
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
		[]task.Filter{},
		&retry.TaskRetry{},
		1,
		3,
		manager.GroupInfo{},
	}}, nil
}

func (m MockBrokenApiManager) Deploy([]byte) ([]*manager.Task, error) {
	return nil, errors.New("Broken")
}
func (m MockBrokenApiManager) Kill([]byte) (string, error) { return "", errors.New("Broken") }
func (m MockBrokenApiManager) Update([]byte) ([]*manager.Task, error) {
	return nil, errors.New("Broken")
}
func (m MockBrokenApiManager) Status(string) (mesos_v1.TaskState, error) {
	return mesos_v1.TaskState_TASK_UNKNOWN, errors.New("Broken")
}
func (m MockBrokenApiManager) AllTasks() ([]manager.Task, error) {
	return nil, errors.New("Broken")
}

package test

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/structures/test"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"sprint/task/retry"
)

type MockTaskManager struct{}

func (m MockTaskManager) AddPolicy(*task.TimeRetry, *mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManager) CheckPolicy(*mesos_v1.TaskInfo) (*retry.TaskRetry, error) {
	return nil, nil
}
func (m MockTaskManager) ClearPolicy(*mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManager) RunPolicy(*retry.TaskRetry, func() error) error {
	return nil
}

func (m MockTaskManager) Add(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManager) Delete(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManager) Get(*string) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockTaskManager) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManager) State(name *string) (*mesos_v1.TaskState, error) {
	return new(mesos_v1.TaskState), nil
}

func (m MockTaskManager) TotalTasks() int {
	return 0
}

func (m MockTaskManager) Tasks() structures.DistributedMap {
	return &test.MockDistributedMap{}
}

func (m MockTaskManager) All() ([]manager.Task, error) {
	return []manager.Task{{
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
	}}, nil
}

func (m MockTaskManager) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return []*mesos_v1.TaskInfo{{}}, nil
}

//
// Mock Broken Task Manager
//
type MockBrokenTaskManager struct{}

func (m MockBrokenTaskManager) AddPolicy(*task.TimeRetry, *mesos_v1.TaskInfo) error {
	return nil
}
func (m MockBrokenTaskManager) CheckPolicy(*mesos_v1.TaskInfo) (*retry.TaskRetry, error) {
	return nil, nil
}
func (m MockBrokenTaskManager) ClearPolicy(*mesos_v1.TaskInfo) error {
	return nil
}
func (m MockBrokenTaskManager) RunPolicy(*retry.TaskRetry, func() error) error {
	return nil
}

func (m MockBrokenTaskManager) Add(*mesos_v1.TaskInfo) error {
	return errors.New("Broken.")
}

func (m MockBrokenTaskManager) Delete(*mesos_v1.TaskInfo) error {
	return errors.New("Broken.")
}

func (m MockBrokenTaskManager) Get(*string) (*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockBrokenTaskManager) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) error {
	return errors.New("Broken.")
}

func (m MockBrokenTaskManager) State(name *string) (*mesos_v1.TaskState, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) TotalTasks() int {
	return 0
}

func (m MockBrokenTaskManager) Tasks() structures.DistributedMap {
	return &test.MockBrokenDistributedMap{}
}

func (m MockBrokenTaskManager) All() ([]manager.Task, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken.")
}

type MockTaskManagerQueued struct{}

func (m MockTaskManagerQueued) AddPolicy(*task.TimeRetry, *mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManagerQueued) CheckPolicy(*mesos_v1.TaskInfo) (*retry.TaskRetry, error) {
	return nil, nil
}
func (m MockTaskManagerQueued) ClearPolicy(*mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManagerQueued) RunPolicy(*retry.TaskRetry, func() error) error {
	return nil
}

func (m MockTaskManagerQueued) Add(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) Delete(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) Get(*string) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManagerQueued) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManagerQueued) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockTaskManagerQueued) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) State(name *string) (*mesos_v1.TaskState, error) {
	return new(mesos_v1.TaskState), nil
}

func (m MockTaskManagerQueued) AllByState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return []*mesos_v1.TaskInfo{{}}, nil
}

func (m MockTaskManagerQueued) TotalTasks() int {
	return 1
}

func (m MockTaskManagerQueued) Tasks() structures.DistributedMap {
	return &test.MockDistributedMap{}
}

func (m MockTaskManagerQueued) All() ([]manager.Task, error) {
	return []manager.Task{{
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
	}}, nil
}

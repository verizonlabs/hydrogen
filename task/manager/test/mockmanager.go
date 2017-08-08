package test

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/structures/test"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/task/retry"
	"time"
)

type MockTaskManager struct{}

func (m MockTaskManager) Add(...*manager.Task) error {
	return nil
}

func (m MockTaskManager) Delete(...*manager.Task) error {
	return nil
}

func (m MockTaskManager) Get(*string) (*manager.Task, error) {
	return &manager.Task{Retry: &retry.TaskRetry{
		TotalRetries: 0,
		MaxRetries:   0,
		RetryTime:    time.Nanosecond,
		Backoff:      true,
		Name:         "id",
	}}, nil
}

func (m MockTaskManager) GetById(id *mesos_v1.TaskID) (*manager.Task, error) {
	return &manager.Task{Retry: &retry.TaskRetry{
		TotalRetries: 0,
		MaxRetries:   0,
		RetryTime:    time.Nanosecond,
		Backoff:      true,
		Name:         "id",
	}}, nil
}

func (m MockTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockTaskManager) Update(...*manager.Task) error {
	return nil
}

func (m MockTaskManager) AllByState(state mesos_v1.TaskState) ([]*manager.Task, error) {
	return []*manager.Task{}, nil
}

func (m MockTaskManager) TotalTasks() int {
	return 0
}

func (m MockTaskManager) All() ([]manager.Task, error) {
	return []manager.Task{*manager.NewTask(
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
		[]task.Filter{},
		&retry.TaskRetry{},
		1,
		manager.GroupInfo{})}, nil
}

//
// Mock Broken Task Manager
//
type MockBrokenTaskManager struct{}

var (
	broken error = errors.New("Broken")
)

func (m MockBrokenTaskManager) Add(...*manager.Task) error {
	return broken
}

func (m MockBrokenTaskManager) Delete(...*manager.Task) error {
	return broken
}

func (m MockBrokenTaskManager) Get(*string) (*manager.Task, error) {
	return nil, broken
}

func (m MockBrokenTaskManager) GetById(id *mesos_v1.TaskID) (*manager.Task, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockBrokenTaskManager) Update(...*manager.Task) error {
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

func (m MockBrokenTaskManager) AllByState(state mesos_v1.TaskState) ([]*manager.Task, error) {
	return nil, errors.New("Broken.")
}

type MockTaskManagerQueued struct{}

func (m MockTaskManagerQueued) CreateGroup(string) error                     { return nil }
func (m MockTaskManagerQueued) Unlink(string, *mesos_v1.AgentID) error       { return nil }
func (m MockTaskManagerQueued) SetSize(string, int) error                    { return nil }
func (m MockTaskManagerQueued) Link(string, *mesos_v1.AgentID) error         { return nil }
func (m MockTaskManagerQueued) AddToGroup(string, *mesos_v1.AgentID) error   { return nil }
func (m MockTaskManagerQueued) ReadGroup(string) []*mesos_v1.AgentID         { return []*mesos_v1.AgentID{} }
func (m MockTaskManagerQueued) DelFromGroup(string, *mesos_v1.AgentID) error { return nil }
func (m MockTaskManagerQueued) DeleteGroup(string) error                     { return nil }
func (m MockTaskManagerQueued) IsInGroup(*mesos_v1.TaskInfo) bool            { return true }

func (m MockTaskManagerQueued) AddPolicy(*task.TimeRetry, *mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManagerQueued) CheckPolicy(*mesos_v1.TaskInfo) *retry.TaskRetry {
	return nil
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
	return []manager.Task{*manager.NewTask(
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
		[]task.Filter{},
		&retry.TaskRetry{},
		1,
		manager.GroupInfo{})}, nil
}

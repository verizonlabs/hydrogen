package retry

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/task"
	"time"
)

type (
	Retry interface {
		AddPolicy(policy *task.TimeRetry, mesosTask *mesos_v1.TaskInfo) error
		CheckPolicy(mesosTask *mesos_v1.TaskInfo) (*TaskRetry, error)
		ClearPolicy(mesosTask *mesos_v1.TaskInfo) error
		RunPolicy(policy *TaskRetry, f func() error) error
	}
	TaskRetry struct {
		TotalRetries int
		MaxRetries   int
		RetryTime    time.Duration
		Backoff      bool
		Name         string
	}
)

func NewDefaultPolicy(name string) *TaskRetry {
	return &TaskRetry{
		TotalRetries: 0,
		RetryTime:    1 * time.Second,
		Backoff:      true,
		Name:         name,
	}
}

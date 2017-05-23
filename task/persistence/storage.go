package persistence

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/persistence"
	"mesos-framework-sdk/task"
	"sprint/scheduler"
	"sprint/task/retry"
	"time"
)

type Storage interface {
	retry.Retry
	persistence.KeyValueStore
}

type Persistence struct {
	persistence.KeyValueStore
	policy retry.TaskRetry
}

func NewPersistence(kv persistence.KeyValueStore, config *scheduler.Configuration) Storage {
	return &Persistence{
		KeyValueStore: kv,
		policy: retry.TaskRetry{
			RetryTime:  time.Duration(2),
			MaxRetries: config.Persistence.MaxRetries,
			Backoff:    true,
		},
	}
}

// We're not taking user input for storage policies.
// Return nil error to satisfy the interface.
func (p Persistence) AddPolicy(policy *task.TimeRetry, mesosTask *mesos_v1.TaskInfo) error {
	return nil
}

func (p Persistence) CheckPolicy(mesosTask *mesos_v1.TaskInfo) (*retry.TaskRetry, error) {
	return &p.policy, nil
}

// We're not storing policies constructed from user input in memory.
// Return nil error to satisfy the interface.
func (p Persistence) ClearPolicy(mesosTask *mesos_v1.TaskInfo) error {
	return nil
}

func (p Persistence) RunPolicy(policy *retry.TaskRetry, f func() error) error {
	if policy.TotalRetries == policy.MaxRetries {
		return errors.New("Retry limit reached")
	}

	err := f()
	policy.TotalRetries += 1
	if err != nil {
		policy.RetryTime = policy.RetryTime * 2
		time.Sleep(policy.RetryTime)
		return p.RunPolicy(policy, f)
	}

	return nil
}

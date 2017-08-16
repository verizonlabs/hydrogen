// Copyright 2017 Verizon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package persistence

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/persistence"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/retry"
	"sprint/scheduler"
	"time"
)

// Provides pluggable storage types that can be used to persist state.
// Also used extensively for testing with mocks.
type Storage interface {
	retry.Retry
	persistence.KeyValueStore
}

// Primary persistence engine that's used to store task state, high availability metadata, and more.
type Persistence struct {
	persistence.KeyValueStore
	policy retry.TaskRetry
}

// Returns the main persistence engine that's used across the framework.
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

// Returns the current policy that's used by the persistence engine.
// There is only ever one policy used at a given time.
func (p Persistence) CheckPolicy(mesosTask *mesos_v1.TaskInfo) *retry.TaskRetry {
	return &p.policy
}

// We're not storing policies constructed from user input in memory.
// Return nil error to satisfy the interface.
func (p Persistence) ClearPolicy(mesosTask *mesos_v1.TaskInfo) error {
	return nil
}

// Runs the supplied policy for storage operations.
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

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

package builder

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	resourcebuilder "mesos-framework-sdk/resources"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/command"
	"mesos-framework-sdk/task/container"
	"mesos-framework-sdk/task/healthcheck"
	"mesos-framework-sdk/task/labels"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/task/resources"
	"mesos-framework-sdk/task/retry"
	"mesos-framework-sdk/utils"
	"time"
)

var NoNameError = errors.New("A name is required for the application. Please set the name field.")
var NoResourcesError = errors.New("Application requested with no resources. Please set some resources.")

// Parses a 1...n tasks.  Any error fails all other tasks.
func Application(tasks ...*task.ApplicationJSON) ([]*manager.Task, error) {
	parsedTasks := []*manager.Task{}

	for _, t := range tasks {
		taskIntent := &manager.Task{}

		// Check for required name.
		if t.Name == "" {
			// Fail
			return nil, NoNameError
		}
		if t.Resources == nil {
			// Fail
			return nil, NoResourcesError
		}

		// Agent and TaskID are required but are set by the Resource Manager in the scheduler.

		// Executor or CommandInfo must be set.
		// An end user will never be allowed to set an executor, only other frameworks.
		// So here we assume a commandInfo.
		cmd, err := command.ParseCommandInfo(t.Command)
		if err != nil {
			// If we don't have a commandInfo, it's invalid.
			return nil, err
		}

		// Parse resources
		// These are required, fail if no resources are specified.
		res, err := resources.ParseResources(t.Resources)
		if err != nil {
			return nil, err
		}

		// Container parse
		image, err := container.ParseContainer(t.Container)
		if err != nil {
			return nil, err
		}

		lbls, err := labels.ParseLabels(t.Labels)
		if err != nil {
			return nil, err
		}

		hc, err := healthcheck.ParseHealthCheck(t.HealthCheck, cmd)
		if err != nil {
			return nil, err
		}

		name := t.Name
		taskId := &mesos_v1.TaskID{Value: utils.ProtoString(name)}

		if len(t.Filters) > 0 {
			taskIntent.Filters = t.Filters
		}

		if t.Retry != nil {
			duration, err := time.ParseDuration(t.Retry.Time + "s")
			if err != nil {
				return nil, err
			}
			taskIntent.Retry = &retry.TaskRetry{
				TotalRetries: 0,
				MaxRetries:   t.Retry.MaxRetries,
				RetryTime:    duration,
				Backoff:      t.Retry.Backoff,
				Name:         t.Name,
			}
		} else {
			// Default retry policy.
			taskIntent.Retry = &retry.TaskRetry{
				TotalRetries: 0,
				MaxRetries:   2,
				RetryTime:    time.Duration(1 * time.Second),
				Backoff:      true,
				Name:         t.Name,
			}
		}

		taskIntent.Instances = t.Instances

		taskIntent.Strategy = t.Strategy

		taskIntent.Info = resourcebuilder.CreateTaskInfo(
			utils.ProtoString(name),
			taskId,
			cmd,
			res,
			image,
			hc,
			lbls,
		)

		parsedTasks = append(parsedTasks, taskIntent)
	}
	return parsedTasks, nil
}

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

package controller

import (
	"github.com/verizonlabs/hydrogen/scheduler/ha"
	"github.com/verizonlabs/hydrogen/task/manager"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler/events"
	sdkTaskManager "github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/plan"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	// Sprint Executor runs plans,
	// or waits until a new plan comes along.
	// The executor never stops unless the process is killed.
	SprintExecutor struct {
		planner plan.PlanQueue
		logger  logging.Logger
	}
)

// Returns the main controller that's used to coordinate the calls/events from/to the scheduler.
func NewSprintExecutor(planner plan.PlanQueue, logger logging.Logger) *SprintExecutor {
	return &SprintExecutor{
		planner: planner,
		logger:  logger,
	}
}

// Run executes the high-level plan logic for Sprint.
func (s *SprintExecutor) Run() {
	var currentPlan plan.Plan
	for {
		// See if we have a plan to run
		currentPlan = s.planner.Peek()

		// If not, just block and wait for a new plan to be put on the stack.
		if currentPlan.Type() == plan.Idle {
			currentPlan = <-s.planner.Wait()
		}

		// Check if the plan is already running
		if currentPlan.Status() != plan.RUNNING {
			// If not, run the plan.
			// Run the current plan once we have one to run that's not idle.
			// The plan gets updated in Execute automatically
			currentPlan.Execute()
		}

		// Check status and make sure it returns successful, if not,
		// loop and check the status again.
		if currentPlan.Status() == plan.FINISHED {
			s.planner.Pop()
		}
	}
}

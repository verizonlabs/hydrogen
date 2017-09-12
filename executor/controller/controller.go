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
	e "mesos-framework-sdk/executor"
	"mesos-framework-sdk/executor/events"
	exec "mesos-framework-sdk/include/mesos_v1_executor"
	"mesos-framework-sdk/logging"
	"time"
)

const (
	subscribeRetry = 2
)

type ExecutorController struct {
	executor  e.Executor
	logger    logging.Logger
	eventChan chan *exec.Event
}

func NewEventController(e e.Executor, l logging.Logger) *ExecutorController {
	return &ExecutorController{
		executor:  e,
		eventChan: make(chan *exec.Event),
		logger:    l,
	}
}

// Run subscribes and starts listening for executor events.
func (d *ExecutorController) Run(e events.ExecutorEvents) {
	go func() {
		for {
			err := d.executor.Subscribe(d.eventChan)
			if err != nil {
				d.logger.Emit(logging.ERROR, "Failed to subscribe: %s", err.Error())
				time.Sleep(time.Duration(subscribeRetry) * time.Second)
			}
		}
	}()

	d.listen(e)
}

// Listens for Mesos events and routes to the appropriate handler.
func (d *ExecutorController) listen(e events.ExecutorEvents) {
	for {
		switch t := <-d.eventChan; t.GetType() {
		case exec.Event_SUBSCRIBED:
			e.Subscribed(t.GetSubscribed())
		case exec.Event_ACKNOWLEDGED:
			e.Acknowledged(t.GetAcknowledged())
		case exec.Event_MESSAGE:
			e.Message(t.GetMessage())
		case exec.Event_KILL:
			e.Kill(t.GetKill())
		case exec.Event_LAUNCH:
			e.Launch(t.GetLaunch())
		case exec.Event_LAUNCH_GROUP:
			e.LaunchGroup(t.GetLaunchGroup())
		case exec.Event_SHUTDOWN:
			e.Shutdown()
		case exec.Event_ERROR:
			e.Error(t.GetError())
		case exec.Event_UNKNOWN:
			d.logger.Emit(logging.ALARM, "Unknown event received")
		}
	}
}

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
	e "github.com/verizonlabs/mesos-framework-sdk/executor"
	"github.com/verizonlabs/mesos-framework-sdk/executor/events"
	exec "github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_executor"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
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
		event := <-d.eventChan
		e.Run(event)
	}
}

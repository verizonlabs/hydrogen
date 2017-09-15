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

package main

import (
	"hydrogen/executor/controller"
	"hydrogen/executor/events"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/executor"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/utils"
	"os"
)

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	logger := logging.NewDefaultLogger()

	// Environment vars are implicitly set and provided by the Mesos agent.
	fwId := &mesos_v1.FrameworkID{Value: utils.ProtoString(os.Getenv("MESOS_FRAMEWORK_ID"))}
	execId := &mesos_v1.ExecutorID{Value: utils.ProtoString(os.Getenv("MESOS_EXECUTOR_ID"))}
	protocol := os.Getenv("PROTOCOL")
	endpoint := protocol + "://" + os.Getenv("MESOS_AGENT_ENDPOINT") + "/api/v1/executor"
	auth := "Bearer " + os.Getenv("MESOS_EXECUTOR_AUTHENTICATION_TOKEN") // Passed to us by the agent.

	c := client.NewClient(client.ClientData{
		Endpoint: endpoint,
		Auth:     auth,
	}, logger)
	ex := executor.NewDefaultExecutor(fwId, execId, c, logger)
	handler := events.NewHandler(ex, logger)

	controller.NewEventController(ex, logger).Run(handler)
}

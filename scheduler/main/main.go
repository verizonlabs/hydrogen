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
	"encoding/base64"
	"flag"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources/manager"
	sched "mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	"mesos-framework-sdk/server/file"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	t "mesos-framework-sdk/task/manager"
	"hydrogen/scheduler"
	"hydrogen/scheduler/api"
	apiManager "hydrogen/scheduler/api/manager"
	"hydrogen/scheduler/controller"
	"hydrogen/scheduler/events"
	"hydrogen/scheduler/ha"
	sprintTaskManager "hydrogen/task/manager"
	"hydrogen/task/persistence"
	"strings"
)

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	logger := logging.NewDefaultLogger()

	// Define our framework here.
	config := new(scheduler.Configuration).Initialize()
	frameworkInfo := &mesos_v1.FrameworkInfo{
		User:            &config.Scheduler.User,
		Name:            &config.Scheduler.Name,
		FailoverTimeout: &config.Scheduler.Failover,
		Checkpoint:      &config.Scheduler.Checkpointing,
		Role:            &config.Scheduler.Role,
		Hostname:        &config.Scheduler.Hostname,
		Principal:       &config.Scheduler.Principal,
	}

	flag.Parse()

	// Executor Server
	execSrvCfg := server.NewConfiguration(
		config.FileServer.Cert,
		config.FileServer.Key,
		config.FileServer.Path,
		config.FileServer.Port,
	)
	executorSrv := file.NewExecutorServer(execSrvCfg, logger)

	// Executor server serves up our custom executor binary, if any.
	logger.Emit(logging.INFO, "Starting executor file server")
	go executorSrv.Serve()

	// Storage interface that holds client and retry policy manager.
	p := persistence.NewPersistence(etcd.NewClient(
		strings.Split(config.Persistence.Endpoints, ","),
		config.Persistence.Timeout,
		config.Persistence.KeepaliveTime,
		config.Persistence.KeepaliveTimeout,
	), config.Persistence.MaxRetries)

	// Manages our tasks.
	taskManager := sprintTaskManager.NewTaskManager(
		make(map[string]*t.Task),
		p,
		logger,
	)

	auth := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(config.Scheduler.Principal+":"+config.Scheduler.Secret),
	)

	r := manager.NewDefaultResourceManager() // Manages resources from the cluster
	c := client.NewClient(client.ClientData{
		Endpoint: config.Scheduler.MesosEndpoint,
		Auth:     auth,
	}, logger) // Manages scheduler/executor HTTP calls, authorization, and new master detection.
	s := sched.NewDefaultScheduler(c, frameworkInfo, logger) // Manages how to route and schedule tasks.
	m := apiManager.NewApiParser(r, taskManager, s)          // Middleware for our API.
	ha := ha.NewHA(p, logger, config.Leader)

	// Used to listen for events coming from mesos master to our scheduler.
	eventChan := make(chan *mesos_v1_scheduler.Event)
	reviveChan := make(chan *sdkTaskManager.Task)

	// Event controller manages scheduler events and how they are handled.
	e := controller.NewEventController(config, s, taskManager, p, logger, ha)

	logger.Emit(logging.INFO, "Starting API server")

	// Run our API in a go routine to listen for user requests.
	config.APIServer.Server = server.NewConfiguration(
		config.APIServer.Cert,
		config.APIServer.Key,
		"",
		config.APIServer.Port,
	)

	apiSrv := api.NewApiServer(config, m, logger)
	go apiSrv.RunAPI(nil) // nil means to use default handlers.

	// Run our event controller and kick off HA leader election.
	// Then subscribe to Mesos and start listening for events.
	h := events.NewHandler(taskManager, r, config, s, p, reviveChan, logger)
	e.Run(eventChan, reviveChan, h)
}

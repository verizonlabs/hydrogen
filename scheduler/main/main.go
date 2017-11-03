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
	"github.com/verizonlabs/hydrogen/scheduler"
	"github.com/verizonlabs/hydrogen/scheduler/api"
	apiManager "github.com/verizonlabs/hydrogen/scheduler/api/manager"
	"github.com/verizonlabs/hydrogen/scheduler/events"
	"github.com/verizonlabs/hydrogen/scheduler/ha"
	"github.com/verizonlabs/hydrogen/task/manager"
	"github.com/verizonlabs/hydrogen/task/persistence"
	"github.com/verizonlabs/mesos-framework-sdk/client"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
	"github.com/verizonlabs/mesos-framework-sdk/persistence/drivers/etcd"
	resourceManager "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	sched "github.com/verizonlabs/mesos-framework-sdk/scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/server"
	"github.com/verizonlabs/mesos-framework-sdk/server/file"
	t "github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/scheduler/controller"
	"hydrogen/task/plan"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
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
	taskManager := manager.NewTaskManager(
		make(map[string]*t.Task),
		p,
		logger,
	)

	auth := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(config.Scheduler.Principal+":"+config.Scheduler.Secret),
	)

	r := resourceManager.NewDefaultResourceManager() // Manages resources from the cluster
	c := client.NewClient(client.ClientData{
		Endpoint: config.Scheduler.MesosEndpoint,
		Auth:     auth,
	}, logger) // Manages scheduler/executor HTTP calls, authorization, and new master detection.

	s := sched.NewDefaultScheduler(c, frameworkInfo, logger) // Manages how to route and schedule tasks.

	ha := ha.NewHA(p, logger, config.Leader)

	// Plan manager handles which current plan is running.
	planQueue := plan.NewPlanQueue()

	eventChan := make(chan *mesos_v1_scheduler.Event)

	// Plan factory produces a new plan.
	planFactory := plan.NewPlanFactory(taskManager,
		r,
		s,
		p,
		ha,
		*config,
		logger,
		eventChan,
	)

	// Create a subscribe plan and push it onto the plan queue.
	subscribePlan := planFactory([]*t.Task{}, plan.Subscribe)
	planQueue.Push(subscribePlan)

	// API SECTION
	logger.Emit(logging.INFO, "Starting API server")
	// Run our API in a go routine to listen for user requests.
	config.APIServer.Server = server.NewConfiguration(
		config.APIServer.Cert,
		config.APIServer.Key,
		"",
		config.APIServer.Port,
	)

	// API server configuration.
	// The API handles adding plans to our Plan Queue.
	parser := apiManager.NewApiParser(planQueue, planFactory)
	apiSrv := api.NewApiServer(config, parser, logger)
	go apiSrv.RunAPI(nil) // nil means to use default handlers.

	// Run our event controller and kick off HA leader election.
	// Then subscribe to Mesos and start listening for events.
	h := events.NewEventRouter(planQueue, logger)

	// This just runs the loop to watch the queue.
	e := controller.NewSprintExecutor(planQueue, logger)
	e.Run()
}

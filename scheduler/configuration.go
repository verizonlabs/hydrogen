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

package scheduler

import (
	"flag"
	"mesos-framework-sdk/server"
	"os/user"
	"time"
)

// Holds configuration for all of our components.
type Configuration struct {
	Persistence *PersistenceConfiguration
	Leader      *LeaderConfiguration
	APIServer   *ApiConfiguration
	FileServer  *FileServerConfiguration
	Scheduler   *SchedulerConfiguration
	Executor    *ExecutorConfiguration
}

type ExecutorConfiguration struct {
	CustomExecutor bool
	URI            string
	Name           string
	Command        string
	Shell          bool
	TLS            bool
}

// Persistence connection configuration.
type PersistenceConfiguration struct {
	Endpoints        string
	Timeout          time.Duration
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration
	MaxRetries       int
}

// Configuration for leader (HA) operation.
type LeaderConfiguration struct {
	IP            string
	ServerPort    int
	AddressFamily string
	RetryInterval time.Duration
	ServerRetry   time.Duration
}

// Holds configuration for the built-in REST API.
type ApiConfiguration struct {
	Server  server.Configuration
	Version string
	Cert    string
	Key     string
	Port    int
}

// Configuration for the file (executor) server.
type FileServerConfiguration struct {
	Cert string
	Key  string
	Port int
	Path string
}

// Configuration for the main scheduler.
type SchedulerConfiguration struct {
	CustomExecutor    bool
	MesosEndpoint     string
	Name              string
	User              string
	Role              string
	Checkpointing     bool
	Principal         string
	Secret            string
	ExecutorSrvCfg    server.Configuration
	ExecutorName      string
	ExecutorCmd       string
	Failover          float64
	Hostname          string
	ReconcileInterval time.Duration
	SubscribeRetry    time.Duration
}

// Stores and initializes all of our configuration.
func (c *Configuration) Initialize() *Configuration {
	return &Configuration{
		Persistence: new(PersistenceConfiguration).initialize(),
		Leader:      new(LeaderConfiguration).initialize(),
		APIServer:   new(ApiConfiguration).initialize(),
		FileServer:  new(FileServerConfiguration).initialize(),
		Scheduler:   new(SchedulerConfiguration).initialize(),
		Executor:    new(ExecutorConfiguration).initialize(),
	}
}

// Configuration for the custom executor
func (c *ExecutorConfiguration) initialize() *ExecutorConfiguration {
	flag.BoolVar(&c.CustomExecutor, "executor.enable", false, "Enable/disable usage of the custom executor")
	flag.StringVar(&c.URI, "executor.uri", "http://127.0.0.1:8081/executor", "URI for Mesos to fetch the custom executor")
	flag.StringVar(&c.Name, "executor.name", "Oxygen", "The executor's name")
	flag.StringVar(&c.Command, "executor.command", "executor", "Command to run the executor")
	flag.BoolVar(&c.Shell, "executor.shell", false, "Whether or not the executor should be launched under a shell")
	flag.BoolVar(&c.TLS, "executor.tls", false, "Use TLS when connecting to Mesos")

	return c
}

// Applies default configuration for our persistence connection.
func (c *PersistenceConfiguration) initialize() *PersistenceConfiguration {
	flag.StringVar(&c.Endpoints, "persistence.endpoints", "http://127.0.0.1:2379", "Comma-separated list of "+
		"storage endpoints")
	flag.DurationVar(&c.Timeout, "persistence.timeout", 2*time.Second, "Timeout for CRUD storage operations")
	flag.DurationVar(&c.KeepaliveTime, "persistence.keepalive.time", 30*time.Second, "After a duration of this time "+
		"if the client doesn't see any activity "+
		"it pings the server to see "+
		"if the transport is still alive")
	flag.DurationVar(&c.KeepaliveTimeout, "persistence.keepalive.timeout", 20*time.Second, "After having pinged for keepalive check, "+
		"the client waits for this duration "+
		"and if no activity is seen "+
		"even after that the connection is closed")
	flag.IntVar(&c.MaxRetries, "persistence.retry.max", 3, "How many times persistence operations will be retried")

	return c
}

// Applies default leader configuration.
func (c *LeaderConfiguration) initialize() *LeaderConfiguration {
	flag.StringVar(&c.IP, "ha.ip", "127.0.0.1", "IP address of the node where this framework is running")
	flag.IntVar(&c.ServerPort, "ha.leader.server.port", 8082, "Port that the leader server listens on")
	flag.StringVar(&c.AddressFamily, "ha.leader.server.address.family", "tcp4", "tcp4, tcp6, or tcp for dual stack")
	flag.DurationVar(&c.RetryInterval, "ha.leader.election.retry", 2*time.Second, "How long to wait before retrying "+
		"the leader election process")
	flag.DurationVar(&c.ServerRetry, "ha.leader.server.retry", 2*time.Second, "How long to wait before accepting "+
		"connections from clients after an error")

	return c
}

// Applies default configuration to our API server.
func (c *ApiConfiguration) initialize() *ApiConfiguration {
	flag.StringVar(&c.Version, "api.server.version", "v1", "The API server version to be used")
	flag.StringVar(&c.Cert, "api.server.cert", "", "API server's TLS certificate")
	flag.StringVar(&c.Key, "api.server.key", "", "API server's TLS key")
	flag.IntVar(&c.Port, "api.server.port", 8080, "API server's port")

	return c
}

// Applies default configuration to our file server.
func (c *FileServerConfiguration) initialize() *FileServerConfiguration {
	flag.StringVar(&c.Cert, "file.server.cert", "", "File server's TLS certificate")
	flag.StringVar(&c.Key, "file.server.key", "", "File server's TLS key")
	flag.IntVar(&c.Port, "file.server.port", 8081, "File server's port")
	flag.StringVar(&c.Path, "file.server.path", "executor", "Path to the executor binary")

	return c
}

// Applies default configuration for our scheduler.
func (c *SchedulerConfiguration) initialize() *SchedulerConfiguration {
	u, err := user.Current()
	if err != nil {
		panic("Unable to detect current user: " + err.Error())
	}

	flag.StringVar(&c.MesosEndpoint, "endpoint", "http://127.0.0.1:5050/api/v1/scheduler", "Mesos scheduler API endpoint")
	flag.StringVar(&c.Name, "name", "Hydrogen", "Framework name")
	flag.StringVar(&c.User, "user", u.Username, "User that the executor/task will be launched as")
	flag.StringVar(&c.Role, "role", "*", "Framework role")
	flag.BoolVar(&c.Checkpointing, "checkpointing", true, "Enable or disable checkpointing")
	flag.StringVar(&c.Principal, "principal", "Hydrogen", "Framework principal")
	flag.StringVar(&c.Secret, "secret", "", "Used when Mesos requires authentication")
	flag.Float64Var(&c.Failover, "failover", 168*time.Hour.Seconds(), "Framework failover timeout") // 1 week is recommended
	flag.StringVar(&c.Hostname, "hostname", "", "The framework's hostname")
	flag.DurationVar(&c.SubscribeRetry, "subscribe.retry", 2*time.Second, "Controls the interval at which subscribe "+
		"calls will be retried")
	flag.DurationVar(&c.ReconcileInterval, "reconcile.interval", 15*time.Minute, "How often periodic reconciling happens")

	return c
}

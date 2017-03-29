package scheduler

import (
	"flag"
	"mesos-framework-sdk/server"
	"os/user"
	"time"
)

// Configuration for the scheduler, populated by user-supplied flags.
type SchedulerConfiguration struct {
	MesosEndpoint           string
	Name                    string
	User                    string
	Role                    string
	Checkpointing           bool
	Principal               string
	ExecutorSrvCfg          server.Configuration
	ExecutorName            string
	ExecutorCmd             string
	StorageEndpoints        string
	StorageTimeout          time.Duration
	StorageKeepaliveTime    time.Duration
	StorageKeepaliveTimeout time.Duration
	Failover                float64
	Hostname                string
	ReconcileInterval       time.Duration
	LeaderIP                string
	LeaderServerPort        int
	LeaderAddressFamily     string
	LeaderRetryInterval     time.Duration
	LeaderServerRetry       time.Duration
	SubscribeRetry          time.Duration
}

// Applies values to the various configurations from user-supplied flags.
func (c *SchedulerConfiguration) Initialize() *SchedulerConfiguration {
	u, err := user.Current()
	if err != nil {
		panic("Unable to detect current user: " + err.Error())
	}

	flag.StringVar(&c.MesosEndpoint, "endpoint", "http://127.0.0.1:5050/api/v1/scheduler", "Mesos scheduler API endpoint")
	flag.StringVar(&c.Name, "name", "Sprint", "Framework name")
	flag.StringVar(&c.User, "user", u.Username, "User that the executor/task will be launched as")
	flag.StringVar(&c.Role, "role", "*", "Framework role")
	flag.BoolVar(&c.Checkpointing, "checkpointing", true, "Enable or disable checkpointing")
	flag.StringVar(&c.Principal, "principal", "Sprint", "Framework principal")
	flag.StringVar(&c.StorageEndpoints, "persistence.endpoints", "http://127.0.0.1:2379", `Comma-separated list of
											       storage endpoints`)
	flag.DurationVar(&c.StorageTimeout, "persistence.timeout", 2*time.Second, "Initial storage system connection timeout")
	flag.DurationVar(&c.StorageKeepaliveTime, "persistence.keepalive.time", 30*time.Second, `After a duration of this time
												 if the client doesn't see any activity
												 it pings the server to see
												 if the transport is still alive`)
	flag.DurationVar(&c.StorageKeepaliveTimeout, "persistence.keepalive.timeout", 20*time.Second, `After having pinged for keepalive check,
												       the client waits for this duration
												       and if no activity is seen
												       even after that the connection is closed`)
	flag.Float64Var(&c.Failover, "failover", time.Minute.Seconds(), "Framework failover timeout")
	flag.StringVar(&c.Hostname, "hostname", "", "The framework's hostname")
	flag.DurationVar(&c.SubscribeRetry, "subscribe.retry", 2*time.Second, `Controls the interval at which subscribe
									       calls will be retried`)
	flag.DurationVar(&c.ReconcileInterval, "reconcile.interval", 15*time.Minute, "How often periodic reconciling happens")
	flag.StringVar(&c.LeaderIP, "ha.leader.ip", "", "IP address of the node where this framework is running")
	flag.IntVar(&c.LeaderServerPort, "ha.leader.server.port", 8082, "Port that the leader server listens on")
	flag.StringVar(&c.LeaderAddressFamily, "ha.leader.address.family", "tcp4", "tcp4, tcp6, or tcp for dual stack")
	flag.DurationVar(&c.LeaderRetryInterval, "ha.leader.retry", 2*time.Second, `How long to wait before retrying
										    the leader election process`)
	flag.DurationVar(&c.LeaderServerRetry, "ha.leader.server.retry", 2*time.Second, `How long to wait before accepting
											 connections from clients after an error`)

	return c
}

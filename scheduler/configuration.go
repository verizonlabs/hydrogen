package scheduler

import (
	"flag"
	"mesos-framework-sdk/server"
	"os/user"
	"time"
)

// Configuration for the scheduler, populated by user-supplied flags.
type SchedulerConfiguration struct {
	MesosEndpoint     string
	Name              string
	User              string
	Role              string
	Checkpointing     bool
	Principal         string
	ExecutorSrvCfg    server.Configuration
	ExecutorName      string
	ExecutorCmd       string
	StorageEndpoints  string
	StorageTimeout    time.Duration
	Failover          float64
	Hostname          string
	ReconcileInterval time.Duration
	NetworkInterface  string
	LeaderServerPort  int
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
	flag.StringVar(&c.StorageEndpoints, "persistence.endpoints", "http://127.0.0.1:2379", "Comma-separated list of storage endpoints")
	flag.DurationVar(&c.StorageTimeout, "persistence.timeout", time.Second, "Storage request timeout")
	flag.Float64Var(&c.Failover, "failover", time.Minute.Seconds(), "Framework failover timeout")
	flag.StringVar(&c.Hostname, "hostname", "", "The framework's hostname")
	flag.DurationVar(&c.ReconcileInterval, "reconcile.interval", 15*time.Minute, "How often periodic reconciling happens")
	flag.StringVar(&c.NetworkInterface, "ha.leader.interface", "", "Interface to use for determining the leader IP")
	flag.IntVar(&c.LeaderServerPort, "ha.leader.server.port", 8082, "Port that the leader server listens on")

	return c
}

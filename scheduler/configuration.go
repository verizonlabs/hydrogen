package scheduler

import (
	"flag"
	"mesos-framework-sdk/server"
	"os/user"
	"time"
)

// Holds configuration for all of our components.
type Configuration struct {
	Persistence *persistenceConfiguration
	Leader      *leaderConfiguration
	Scheduler   *schedulerConfiguration
}

// Persistence connection configuration.
type persistenceConfiguration struct {
	Endpoints        string
	Timeout          time.Duration
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration
	RetryInterval    time.Duration
}

// Configuration for leader (HA) operation.
type leaderConfiguration struct {
	IP            string
	ServerPort    int
	AddressFamily string
	RetryInterval time.Duration
	ServerRetry   time.Duration
}

// Configuration for the main scheduler.
type schedulerConfiguration struct {
	MesosEndpoint     string
	Name              string
	User              string
	Role              string
	Checkpointing     bool
	Principal         string
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
		Persistence: new(persistenceConfiguration).initialize(),
		Leader:      new(leaderConfiguration).initialize(),
		Scheduler:   new(schedulerConfiguration).initialize(),
	}
}

// Applies default configuration for our persistence connection.
func (c *persistenceConfiguration) initialize() *persistenceConfiguration {
	flag.StringVar(&c.Endpoints, "persistence.endpoints", "http://127.0.0.1:2379", `Comma-separated list of
											       storage endpoints`)
	flag.DurationVar(&c.Timeout, "persistence.timeout", 2*time.Second, "Initial storage system connection timeout")
	flag.DurationVar(&c.KeepaliveTime, "persistence.keepalive.time", 30*time.Second, `After a duration of this time
												 if the client doesn't see any activity
												 it pings the server to see
												 if the transport is still alive`)
	flag.DurationVar(&c.KeepaliveTimeout, "persistence.keepalive.timeout", 20*time.Second, `After having pinged for keepalive check,
												       the client waits for this duration
												       and if no activity is seen
												       even after that the connection is closed`)
	flag.DurationVar(&c.RetryInterval, "persistence.retry.interval", 2*time.Second, `How long to wait before
												retrying persistence operations`)

	return c
}

// Applies default leader configuration.
func (c *leaderConfiguration) initialize() *leaderConfiguration {
	flag.StringVar(&c.IP, "ha.leader.ip", "", "IP address of the node where this framework is running")
	flag.IntVar(&c.ServerPort, "ha.leader.server.port", 8082, "Port that the leader server listens on")
	flag.StringVar(&c.AddressFamily, "ha.leader.address.family", "tcp4", "tcp4, tcp6, or tcp for dual stack")
	flag.DurationVar(&c.RetryInterval, "ha.leader.retry", 2*time.Second, `How long to wait before retrying
										    the leader election process`)
	flag.DurationVar(&c.ServerRetry, "ha.leader.server.retry", 2*time.Second, `How long to wait before accepting
											 connections from clients after an error`)

	return c
}

// Applies default configuration for our scheduler.
func (c *schedulerConfiguration) initialize() *schedulerConfiguration {
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
	flag.Float64Var(&c.Failover, "failover", 168*time.Hour.Seconds(), "Framework failover timeout") // 1 week is recommended
	flag.StringVar(&c.Hostname, "hostname", "", "The framework's hostname")
	flag.DurationVar(&c.SubscribeRetry, "subscribe.retry", 2*time.Second, `Controls the interval at which subscribe
									       calls will be retried`)
	flag.DurationVar(&c.ReconcileInterval, "reconcile.interval", 15*time.Minute, "How often periodic reconciling happens")

	return c
}

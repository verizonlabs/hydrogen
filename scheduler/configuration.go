package scheduler

import (
	"flag"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/server"
	"os/user"
	"time"
)

type configuration interface {
	Initialize() *SchedulerConfiguration
	Name() string
	User() string
	Checkpointing() *bool
	Principal() string
	Command() *string
	Uris() []mesos_v1.CommandInfo_URI
	Endpoint() string
	ReviveBurst() int
	ReviveWait() time.Duration
	MaxRefuse() time.Duration
	SetExecutorSrvCfg(server.Configuration) *SchedulerConfiguration
	ExecutorSrvCfg() server.Configuration
	ExecutorName() *string
	ExecutorCmd() *string
	Executors() int
	PersistenceTimeout() time.Duration
	PersistenceEndpoints() string
}

// Configuration for the scheduler, populated by user-supplied flags.
type SchedulerConfiguration struct {
	endpoint       string
	name           string
	user           string
	checkpointing  bool
	principal      string
	uris           []mesos_v1.CommandInfo_URI
	command        string
	reviveBurst    int
	reviveWait     time.Duration
	maxRefuse      time.Duration
	executorSrvCfg server.Configuration
	executorName   string
	executorCmd    string
	executors      int
	endpoints      string
	timeout        time.Duration
}

// Applies values to the various configurations from user-supplied flags.
func (c *SchedulerConfiguration) Initialize() *SchedulerConfiguration {
	u, err := user.Current()
	if err != nil {
		panic("Unable to detect current user: " + err.Error())
	}

	flag.StringVar(&c.endpoint, "endpoint", "http://127.0.0.1:5050/api/v1/scheduler", "Mesos scheduler API endpoint")
	flag.StringVar(&c.name, "name", "Sprint", "Framework name")
	flag.StringVar(&c.user, "user", u.Username, "User that the executor/task will be launched as")
	flag.BoolVar(&c.checkpointing, "checkpointing", true, "Enable or disable checkpointing")
	flag.StringVar(&c.principal, "principal", "Sprint", "Framework principal")
	flag.StringVar(&c.command, "command", "", "Executor command")
	flag.IntVar(&c.reviveBurst, "revive.burst", 3, "Number of revive messages that may be sent in a burst within revive-wait period")
	flag.DurationVar(&c.reviveWait, "revive.wait", 1*time.Second, "Wait this long to fully recharge revive-burst quota")
	flag.DurationVar(&c.maxRefuse, "maxRefuse", 5*time.Second, "Max length of time to refuse future offers")
	flag.StringVar(&c.executorName, "executor.name", "Sprinter", "Name of the executor")
	flag.StringVar(&c.executorCmd, "executor.command", "./executor", "Executor command")
	flag.IntVar(&c.executors, "executor.count", 5, "Number of executors to run.")
	flag.StringVar(&c.endpoints, "persistence.endpoints", "http://127.0.0.1:2379", "Comma-separated list of storage endpoints")
	flag.DurationVar(&c.timeout, "persistence.timeout", time.Second, "Storage request timeout")

	return c
}

func (c *SchedulerConfiguration) Name() string {
	return c.name
}

func (c *SchedulerConfiguration) User() string {
	return c.user
}

func (c *SchedulerConfiguration) Checkpointing() *bool {
	return &c.checkpointing
}

func (c *SchedulerConfiguration) Principal() string {
	return c.principal
}

func (c *SchedulerConfiguration) Command() *string {
	return &c.command
}

func (c *SchedulerConfiguration) Uris() []mesos_v1.CommandInfo_URI {
	return c.uris
}

func (c *SchedulerConfiguration) Endpoint() string {
	return c.endpoint
}

func (c *SchedulerConfiguration) ReviveBurst() int {
	return c.reviveBurst
}

func (c *SchedulerConfiguration) ReviveWait() time.Duration {
	return c.reviveWait
}

func (c *SchedulerConfiguration) MaxRefuse() time.Duration {
	return c.maxRefuse
}

func (c *SchedulerConfiguration) SetExecutorSrvCfg(cfg server.Configuration) *SchedulerConfiguration {
	c.executorSrvCfg = cfg

	return c
}

func (c *SchedulerConfiguration) ExecutorSrvCfg() server.Configuration {
	return c.executorSrvCfg
}

func (c *SchedulerConfiguration) ExecutorName() *string {
	return &c.executorName
}

func (c *SchedulerConfiguration) ExecutorCmd() *string {
	return &c.executorCmd
}

func (c *SchedulerConfiguration) Executors() int {
	return c.executors
}

func (c *SchedulerConfiguration) PersistenceTimeout() time.Duration {
	return c.timeout
}

func (c *SchedulerConfiguration) PersistenceEndpoints() string {
	return c.endpoints
}

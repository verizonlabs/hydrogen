package scheduler

import (
	"flag"
	"mesos-sdk"
	"os/user"
	"sprint/scheduler/server"
	"time"
)

type configuration interface {
	Initialize() *SprintConfiguration
	Name() string
	User() string
	Checkpointing() *bool
	Principal() string
	Command() *string
	Uris() []mesos.CommandInfo_URI
	Endpoint() string
	Timeout() time.Duration
	ReviveBurst() int
	ReviveWait() time.Duration
	MaxRefuse() time.Duration
	SetExecutorSrvCfg(server.Configuration) *SprintConfiguration
	ExecutorSrvCfg() server.Configuration
	ExecutorName() *string
	ExecutorCmd() *string
}

// Configuration for the scheduler, populated by user-supplied flags.
type SprintConfiguration struct {
	endpoint       string
	name           string
	user           string
	checkpointing  bool
	principal      string
	uris           []mesos.CommandInfo_URI
	command        string
	timeout        time.Duration
	reviveBurst    int
	reviveWait     time.Duration
	maxRefuse      time.Duration
	executorSrvCfg server.Configuration
	executorName   string
	executorCmd    string
}

// Applies values to the various configurations from user-supplied flags.
func (c *SprintConfiguration) Initialize() *SprintConfiguration {
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
	flag.DurationVar(&c.timeout, "timeout", 20*time.Second, "Mesos connection timeout")
	flag.IntVar(&c.reviveBurst, "revive.burst", 3, "Number of revive messages that may be sent in a burst within revive-wait period")
	flag.DurationVar(&c.reviveWait, "revive.wait", 1*time.Second, "Wait this long to fully recharge revive-burst quota")
	flag.DurationVar(&c.maxRefuse, "maxRefuse", 5*time.Second, "Max length of time to refuse future offers")
	flag.StringVar(&c.executorName, "executor.name", "Sprinter", "Name of the executor")
	flag.StringVar(&c.executorCmd, "executor.command", "./executor", "Executor command")

	return c
}

func (c *SprintConfiguration) Name() string {
	return c.name
}

func (c *SprintConfiguration) User() string {
	return c.user
}

func (c *SprintConfiguration) Checkpointing() *bool {
	return &c.checkpointing
}

func (c *SprintConfiguration) Principal() string {
	return c.principal
}

func (c *SprintConfiguration) Command() *string {
	return &c.command
}

func (c *SprintConfiguration) Uris() []mesos.CommandInfo_URI {
	return c.uris
}

func (c *SprintConfiguration) Timeout() time.Duration {
	return c.timeout
}

func (c *SprintConfiguration) Endpoint() string {
	return c.endpoint
}

func (c *SprintConfiguration) ReviveBurst() int {
	return c.reviveBurst
}

func (c *SprintConfiguration) ReviveWait() time.Duration {
	return c.reviveWait
}

func (c *SprintConfiguration) MaxRefuse() time.Duration {
	return c.maxRefuse
}

func (c *SprintConfiguration) SetExecutorSrvCfg(cfg server.Configuration) *SprintConfiguration {
	c.executorSrvCfg = cfg

	return c
}

func (c *SprintConfiguration) ExecutorSrvCfg() server.Configuration {
	return c.executorSrvCfg
}

func (c *SprintConfiguration) ExecutorName() *string {
	return &c.executorName
}

func (c *SprintConfiguration) ExecutorCmd() *string {
	return &c.executorCmd
}

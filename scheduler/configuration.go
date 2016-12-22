package scheduler

import (
	"flag"
	"mesos-sdk"
	"time"
)

type configuration interface {
	Initialize(fs *flag.FlagSet) *SprintConfiguration
	Name() string
	Checkpointing() *bool
	Principal() string
	Command() *string
	Uris() []mesos.CommandInfo_URI
	Endpoint() string
	Timeout() time.Duration
	ReviveBurst() int
	ReviveWait() time.Duration
	MaxRefuse() time.Duration
}

// Configuration for the scheduler, populated by user-supplied flags.
type SprintConfiguration struct {
	endpoint      string
	name          string
	checkpointing bool
	principal     string
	uris          []mesos.CommandInfo_URI
	command       string
	timeout       time.Duration
	reviveBurst   int
	reviveWait    time.Duration
	maxRefuse     time.Duration
}

// Applies values to the various configurations from user-supplied flags.
func (c *SprintConfiguration) Initialize(fs *flag.FlagSet) *SprintConfiguration {
	fs.StringVar(&c.endpoint, "endpoint", "http://127.0.0.1:5050/api/v1/scheduler", "Mesos scheduler API endpoint")
	fs.StringVar(&c.name, "name", "Sprint", "Framework name")
	fs.BoolVar(&c.checkpointing, "checkpointing", true, "Enable or disable checkpointing")
	fs.StringVar(&c.principal, "principal", "Sprint", "Framework principal")
	fs.StringVar(&c.command, "command", "", "Executor command")
	fs.DurationVar(&c.timeout, "timeout", 20*time.Second, "Mesos connection timeout")
	fs.IntVar(&c.reviveBurst, "revive.burst", 3, "Number of revive messages that may be sent in a burst within revive-wait period")
	fs.DurationVar(&c.reviveWait, "revive.wait", 1*time.Second, "Wait this long to fully recharge revive-burst quota")
	fs.DurationVar(&c.maxRefuse, "maxRefuse", 5*time.Second, "Max length of time to refuse future offers")

	return c
}

func (c *SprintConfiguration) Name() string {
	return c.name
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

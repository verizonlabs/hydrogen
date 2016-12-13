package scheduler

import (
	"flag"
	"github.com/verizonlabs/mesos-go"
	"time"
)

type configuration interface {
	Initialize(fs *flag.FlagSet)
	GetName() string
	GetCheckpointing() *bool
	GetPrincipal() string
	GetCommand() string
	GetUris() []mesos.CommandInfo_URI
	GetEndpoint() string
	GetTimeout() time.Duration
	GetReviveBurst() int
	GetReviveWait() time.Duration
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
}

// Applies values to the various configurations from user-supplied flags.
func (c *SprintConfiguration) Initialize(fs *flag.FlagSet) {
	fs.StringVar(&c.endpoint, "endpoint", "http://127.0.0.1:5050/api/v1/scheduler", "Mesos scheduler API endpoint")
	fs.StringVar(&c.name, "name", "Sprint", "Framework name")
	fs.BoolVar(&c.checkpointing, "checkpointing", true, "Enable or disable checkpointing")
	fs.StringVar(&c.principal, "principal", "Sprint", "Framework principal")
	fs.StringVar(&c.command, "command", "", "Executor command")
	fs.DurationVar(&c.timeout, "timeout", 20*time.Second, "Mesos connection timeout")
	fs.IntVar(&c.reviveBurst, "revive.burst", 3, "Number of revive messages that may be sent in a burst within revive-wait period")
	fs.DurationVar(&c.reviveWait, "revive.wait", 1*time.Second, "Wait this long to fully recharge revive-burst quota")
}

func (c *SprintConfiguration) GetName() string {
	return c.name
}

func (c *SprintConfiguration) GetCheckpointing() *bool {
	return &c.checkpointing
}

func (c *SprintConfiguration) GetPrincipal() string {
	return c.principal
}

func (c *SprintConfiguration) GetCommand() string {
	return c.command
}

func (c *SprintConfiguration) GetUris() []mesos.CommandInfo_URI {
	return c.uris
}

func (c *SprintConfiguration) GetTimeout() time.Duration {
	return c.timeout
}

func (c *SprintConfiguration) GetEndpoint() string {
	return c.endpoint
}

func (c *SprintConfiguration) GetReviveBurst() int {
	return c.reviveBurst
}

func (c *SprintConfiguration) GetReviveWait() time.Duration {
	return c.reviveWait
}

package scheduler

import (
	"flag"
	"github.com/verizonlabs/mesos-go"
	"time"
)

type baseConfiguration interface {
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
type Configuration struct {
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
func (c *Configuration) Initialize(fs *flag.FlagSet) {
	fs.StringVar(&c.endpoint, "endpoint", "http://127.0.0.1:5050/api/v1/scheduler", "Mesos scheduler API endpoint")
	fs.StringVar(&c.name, "name", "Sprint", "Framework name")
	fs.BoolVar(&c.checkpointing, "checkpointing", true, "Enable or disable checkpointing")
	fs.StringVar(&c.principal, "principal", "Sprint", "Framework principal")
	fs.StringVar(&c.command, "command", "", "Executor command")
	fs.DurationVar(&c.timeout, "timeout", 20*time.Second, "Mesos connection timeout")
	fs.IntVar(&c.reviveBurst, "revive.burst", 3, "Number of revive messages that may be sent in a burst within revive-wait period")
	fs.DurationVar(&c.reviveWait, "revive.wait", 1*time.Second, "Wait this long to fully recharge revive-burst quota")
}

func (c *Configuration) GetName() string {
	return c.name
}

func (c *Configuration) GetCheckpointing() *bool {
	return &c.checkpointing
}

func (c *Configuration) GetPrincipal() string {
	return c.principal
}

func (c *Configuration) GetCommand() string {
	return c.command
}

func (c *Configuration) GetUris() []mesos.CommandInfo_URI {
	return c.uris
}

func (c *Configuration) GetTimeout() time.Duration {
	return c.timeout
}

func (c *Configuration) GetEndpoint() string {
	return c.endpoint
}

func (c *Configuration) GetReviveBurst() int {
	return c.reviveBurst
}

func (c *Configuration) GetReviveWait() time.Duration {
	return c.reviveWait
}

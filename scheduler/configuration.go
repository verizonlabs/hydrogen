package scheduler

import (
	"flag"
	"github.com/verizonlabs/mesos-go"
	"time"
)

// Configuration for the scheduler, populated by user-supplied flags.
type Configuration struct {
	endpoint      string
	name          string
	checkpointing bool
	principal     string
	uris          []mesos.CommandInfo_URI
	command       string
	timeout       time.Duration
	reviveBurst   uint
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
	fs.UintVar(&c.reviveBurst, "revive.burst", 3, "Number of revive messages that may be sent in a burst within revive-wait period")
	fs.DurationVar(&c.reviveWait, "revive.wait", 1*time.Second, "Wait this long to fully recharge revive-burst quota")
}

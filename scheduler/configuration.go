package scheduler

import (
	"flag"
	"github.com/verizonlabs/mesos-go"
	"time"
)

type Configuration struct {
	masterEndpoint string
	name           string
	checkpointing  bool
	principal      string
	uris           []mesos.CommandInfo_URI
	command        string
	timeout        time.Duration
}

func (c *Configuration) Initialize(fs *flag.FlagSet) {
	fs.StringVar(&c.masterEndpoint, "masterEndpoint", "http://127.0.0.1:5050", "Mesos master endpoint")
	fs.StringVar(&c.name, "name", "Sprint", "Framework name")
	fs.BoolVar(&c.checkpointing, "checkpointing", true, "Enable or disable checkpointing")
	fs.StringVar(&c.principal, "principal", "Sprint", "Framework principal")
	fs.StringVar(&c.command, "command", "", "Executor command")
	fs.DurationVar(&c.timeout, "timeout", 20*time.Second, "Mesos connection timeout")
}

package executor

import (
	"flag"
	"mesos-sdk/executor/config"
	"time"
)

type Configuration struct {
	internalConfig config.Config
	endpoint       string
	timeout        time.Duration
}

// Create a default configuration if no flags are passed
func (c *Configuration) Initialize(flags *flag.FlagSet) *Configuration {
	flags.StringVar(&c.endpoint, "executor.endpoint", "/api/v1/executor", "Agent API endpoint")
	flags.DurationVar(&c.timeout, "executor.timeout", 10*time.Second, "HTTP timeout")
	flags.StringVar(&c.internalConfig.Directory, "executor.workdir", "slave/", "Working directory of the agent.")
	flags.StringVar(&c.internalConfig.AgentEndpoint, "executor.agent.endpoint", "127.0.0.1:5051", "Endpoint to hit for the agent.")
	flags.DurationVar(&c.internalConfig.ExecutorShutdownGracePeriod, "executor.shutdown.grace.period", 60*time.Second, "Amount of time the agent would wait for an executor to shut down")
	flags.BoolVar(&c.internalConfig.Checkpoint, "executor.checkpoint", true, "Checkpoint is set i.e. when framework checkpointing is enabled; if set then RecoveryTimeout and SubscriptionBackoffMax are also set.")
	flags.DurationVar(&c.internalConfig.RecoveryTimeout, "executor.recovery.timeout", 30*time.Second, "The total duration that the executor should spend retrying before shutting itself down during disconnection.")
	flags.DurationVar(&c.internalConfig.SubscriptionBackoffMax, "executor.subscription.backoff.max", 5*time.Second, "maximum backoff duration to be used by the executor between two retries when disconnected.")

	return c
}

func (c *Configuration) Endpoint() string {
	return c.endpoint
}

func (c *Configuration) Timeout() time.Duration {
	return c.timeout
}

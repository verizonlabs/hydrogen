package execConfig

import (
	"flag"
	"mesos-sdk/executor/config"
	"time"
)

// Create a default configuration if no flags are passed
func Initialize(flags *flag.FlagSet) (config config.Config) {
 	c := config
	flags.StringVar(&c.FrameworkID, "framework.id", "012345-678910", "FrameworkID to be used.")
	flags.StringVar(&c.ExecutorID, "executor.id", "109876-54321", "ExecutorID to be used.")
	flags.StringVar(&c.Directory, "directory", "slave/", "Working directory of the slave.")
	flags.StringVar(&c.AgentEndpoint, "agent.endpoint", "127.0.0.1:5051", "Endpoint to hit for the slave.")
	flags.DurationVar(&c.ExecutorShutdownGracePeriod, "executor.shutdown.grace.period", 60 * time.Second, "Amount of time the slave would wait for an executor to shut down")
	flags.BoolVar(&c.Checkpoint, "checkpoint", true, "Checkpoint is set i.e. when framework checkpointing is enabled; if set then RecoveryTimeout and SubscriptionBackoffMax are also set.")
	flags.DurationVar(&c.RecoveryTimeout, "recovery.timeout", 30 * time.Second, "The total duration that the executor should spend retrying before shutting itself down during disconnection.")
	flags.DurationVar(&c.SubscriptionBackoffMax, "subscription.backoff.max", 5 * time.Second, "maximum backoff duration to be used by the executor between two retries when disconnected.")
	return c
}

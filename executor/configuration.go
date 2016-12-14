package executor

import (
	"mesos-sdk/executor/config"
	"time"
)

type Configuration struct {
	ExecutorConfig config.Config
	ApiEndpoint    string
	Timeout        time.Duration
}

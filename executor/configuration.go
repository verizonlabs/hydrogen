package executor

import (
	"mesos-sdk/executor/config"
	"time"
)

type configuration struct {
	executorConfig config.Config
	apiEndpoint    string
	timeout        time.Duration
}

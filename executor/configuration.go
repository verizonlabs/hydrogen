package executor

import (
	"github.com/verizonlabs/mesos-go/executor/config"
	"time"
)

type configuration struct {
	executorConfig config.Config
	apiEndpoint    string
	timeout        time.Duration
}

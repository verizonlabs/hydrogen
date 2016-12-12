package executor

import (
	"github.com/verizonlabs/mesos-go"
	"github.com/verizonlabs/mesos-go/encoding"
	exec "github.com/verizonlabs/mesos-go/executor"
	"github.com/verizonlabs/mesos-go/executor/calls"
	"github.com/verizonlabs/mesos-go/executor/events"
	"github.com/verizonlabs/mesos-go/httpcli"
	"time"
)

type executor struct {
	config         *configuration
	http           *httpcli.Client
	options        exec.CallOptions
	unackedTasks   map[mesos.TaskID]mesos.TaskInfo
	unackedUpdates map[string]exec.Call_Update
	failedTasks    map[mesos.TaskID]mesos.TaskStatus
	subscribe      *exec.Call
	reconnect      <-chan struct{}
	disconnected   time.Time
	eventHandler   events.Handler
	shutdown       bool
}

func NewExecutor(cfg *configuration) *executor {
	return &executor{
		config: cfg,
		http: httpcli.New(
			httpcli.Endpoint(cfg.apiEndpoint),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.timeout),
				),
			),
		),
		options: exec.CallOptions{
			calls.Framework(cfg.executorConfig.FrameworkID),
			calls.Executor(cfg.executorConfig.ExecutorID),
		},
		unackedTasks:   make(map[mesos.TaskID]mesos.TaskInfo),
		unackedUpdates: make(map[string]exec.Call_Update),
		failedTasks:    make(map[mesos.TaskID]mesos.TaskStatus),
		eventHandler:   NewExecutorHandler(),
	}
}

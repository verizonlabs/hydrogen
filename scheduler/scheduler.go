package scheduler

import (
	"github.com/gogo/protobuf/proto"
	"github.com/verizonlabs/mesos-go"
	"github.com/verizonlabs/mesos-go/encoding"
	"github.com/verizonlabs/mesos-go/httpcli"
	"github.com/verizonlabs/mesos-go/httpcli/httpsched"
	"github.com/verizonlabs/mesos-go/scheduler/calls"
	"net/http"
	"time"
)

type scheduler struct {
	config    *configuration
	framework *mesos.FrameworkInfo
	executor  *mesos.ExecutorInfo
	http      calls.Caller
	shutdown  chan struct{}
}

func NewScheduler(cfg *configuration) *scheduler {
	return &scheduler{
		config: cfg,
		framework: &mesos.FrameworkInfo{
			Name:       cfg.name,
			Checkpoint: &cfg.checkpointing,
		},
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{
				Value: "default",
			},
			Name: proto.String("Sprinter"),
			Command: mesos.CommandInfo{
				Value: proto.String(cfg.command),
				URIs:  cfg.uris,
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_MESOS.Enum(),
			},
		},
		http: httpsched.NewCaller(httpcli.New(
			httpcli.Endpoint(cfg.masterEndpoint),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.timeout),
					httpcli.Transport(func(t *http.Transport) {
						t.ResponseHeaderTimeout = 15 * time.Second
						t.MaxIdleConnsPerHost = 2
					}),
				),
			),
		)),
	}
}

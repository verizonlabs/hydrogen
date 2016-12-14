package main

import (
	"github.com/pborman/uuid"
	"log"
	"mesos-sdk/executor/config"
	"os"
	"sprint/executor"
	"time"
)

func main() {
	cfg := executor.Configuration{
		ExecutorConfig: config.Config{
			FrameworkID:                 "012345",
			ExecutorID:                  uuid.NewRandom().String(),
			Directory:                   "/Users/hansti2/go/src/sprint/executor",
			AgentEndpoint:               "http://127.0.0.1:5051/api/v1/executor",
			ExecutorShutdownGracePeriod: 30 * time.Second,
			RecoveryTimeout:             30 * time.Second,
			SubscriptionBackoffMax:      1 * time.Second,
		},
		Timeout: 20 * time.Second,
	}

	log.Printf("Configuration loaded: %+v", cfg)

	sprint := executor.NewExecutor(&cfg)
	sprint.Subscribe()
	sprint.Run()
	os.Exit(0)
}

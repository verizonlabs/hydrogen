package executor

import (
	"mesos-sdk/executor/config"
	"time"
)

type mockConfiguration struct {
	internalConfig config.Config
	timeout        time.Duration
}

func (m *mockConfiguration) Endpoint() string {
	return ""
}

func (m *mockConfiguration) Timeout() time.Duration {
	return m.timeout
}

func (m *mockConfiguration) SubscriptionBackoffMax() time.Duration {
	return 5 * time.Second
}

func (m *mockConfiguration) Checkpoint() bool {
	return false
}

func (m *mockConfiguration) RecoveryTimeout() time.Duration {
	return 30 * time.Second
}

var cfg configuration = &mockConfiguration{
	timeout: 10 * time.Second,
}

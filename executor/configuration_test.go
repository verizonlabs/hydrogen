package executor

import (
	"mesos-sdk/executor/config"
	"reflect"
	"testing"
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
var executorCfg *ExecutorConfiguration = new(ExecutorConfiguration).Initialize()

// Make sure we can initialize our configuration with default values.
func TestExecutorConfiguration_Initialize(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(executorCfg) != reflect.TypeOf(new(ExecutorConfiguration)) {
		t.Fatal("Executor configuration is not of the right type")
	}
}

// Ensure we get the checkpoint value correctly.
func TestExecutorConfiguration_Checkpoint(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(executorCfg.Checkpoint()) != reflect.TypeOf(*new(bool)) {
		t.Fatal("Checkpoint is of the wrong type")
	}
	if !executorCfg.Checkpoint() {
		t.Fatal("Getter for checkpoint doesn't return the right value")
	}
}

// Measure performance of getting the checkpoint value.
func BenchmarkExecutorConfiguration_Checkpoint(b *testing.B) {
	for n := 0; n < b.N; n++ {
		executorCfg.Checkpoint()
	}
}

// Ensure we get the endpoint correctly.
func TestExecutorConfiguration_Endpoint(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(executorCfg.Endpoint()) != reflect.TypeOf(*new(string)) {
		t.Fatal("Endpoint is of the wrong type")
	}
	if executorCfg.Endpoint() != "http://127.0.0.1:5051/api/v1/executor" {
		t.Fatal("Getter for endpoint doesn't return the right value")
	}
}

// Measure performance of getting our endpoint.
func BenchmarkExecutorConfiguration_Endpoint(b *testing.B) {
	for n := 0; n < b.N; n++ {
		executorCfg.Endpoint()
	}
}

// Ensure we get the recovery timeout correctly.
func TestExecutorConfiguration_RecoveryTimeout(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(executorCfg.RecoveryTimeout()) != reflect.TypeOf(*new(time.Duration)) {
		t.Fatal("Recovery timeout is of the wrong type")
	}
	if executorCfg.RecoveryTimeout() != 30*time.Second {
		t.Fatal("Wrong default value for recovery timeout")
	}
}

// Measure performance of getting our recovery timeout.
func BenchmarkExecutorConfiguration_RecoveryTimeout(b *testing.B) {
	for n := 0; n < b.N; n++ {
		executorCfg.RecoveryTimeout()
	}
}

// Ensure we get the timeout correctly.
func TestExecutorConfiguration_Timeout(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(executorCfg.Timeout()) != reflect.TypeOf(*new(time.Duration)) {
		t.Fatal("Timeout is of the wrong type")
	}
	if executorCfg.Timeout() != 10*time.Second {
		t.Fatal("Wrong default value for timeout")
	}
}

// Measure performance of getting our timeout.
func BenchmarkExecutorConfiguration_Timeout(b *testing.B) {
	for n := 0; n < b.N; n++ {
		executorCfg.Timeout()
	}
}

// Ensure we get the subscription backoff correctly.
func TestExecutorConfiguration_SubscriptionBackoffMax(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(executorCfg.SubscriptionBackoffMax()) != reflect.TypeOf(*new(time.Duration)) {
		t.Fatal("Subscription backoff is of the wrong type")
	}
	if executorCfg.SubscriptionBackoffMax() != 5*time.Second {
		t.Fatal("Wrong default value for subscription backoff")
	}
}

// Measure performance of getting our subscription backoff.
func BenchmarkExecutorConfiguration_SubscriptionBackoffMax(b *testing.B) {
	for n := 0; n < b.N; n++ {
		executorCfg.SubscriptionBackoffMax()
	}
}

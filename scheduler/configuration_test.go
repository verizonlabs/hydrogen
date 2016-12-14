package scheduler

import (
	"flag"
	"mesos-sdk"
	"testing"
	"time"
)

// Mocked configuration
type mockConfiguration struct {
	cfg SprintConfiguration
}

func (m *mockConfiguration) Initialize(fs *flag.FlagSet) *SprintConfiguration {
	m.cfg.name = "Sprint"
	m.cfg.checkpointing = true
	m.cfg.command = ""
	m.cfg.endpoint = "http://127.0.0.1:5050/api/v1/scheduler"
	m.cfg.principal = "Sprint"
	m.cfg.reviveBurst = 3
	m.cfg.reviveWait = 1 * time.Second
	m.cfg.timeout = 20 * time.Second

	return &m.cfg
}

func (m *mockConfiguration) Name() string {
	return m.cfg.name
}

func (m *mockConfiguration) Checkpointing() *bool {
	return &m.cfg.checkpointing
}

func (m *mockConfiguration) Principal() string {
	return m.cfg.principal
}

func (m *mockConfiguration) Command() *string {
	return &m.cfg.command
}

func (m *mockConfiguration) Uris() []mesos.CommandInfo_URI {
	return m.cfg.uris
}

func (m *mockConfiguration) Timeout() time.Duration {
	return m.cfg.timeout
}

func (m *mockConfiguration) Endpoint() string {
	return m.cfg.endpoint
}

func (m *mockConfiguration) ReviveBurst() int {
	return m.cfg.reviveBurst
}

func (m *mockConfiguration) ReviveWait() time.Duration {
	return m.cfg.reviveWait
}

var cfg configuration

func init() {
	cfg = new(mockConfiguration)
}

// Tests setting up default configuration values
func TestConfiguration_Initialize(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)

	config := new(SprintConfiguration).Initialize(fs)

	if config.endpoint != cfg.Endpoint() {
		t.Fatal("Invalid endpoint")
	}
	if config.checkpointing != *cfg.Checkpointing() {
		t.Fatal("Checkpointing is disabled")
	}
	if config.command != *cfg.Command() {
		t.Fatal("Invalid command")
	}
	if config.name != cfg.Name() {
		t.Fatal("Invalid framework name")
	}
	if config.principal != cfg.Principal() {
		t.Fatal("Invalid framework principal")
	}
	if config.timeout != cfg.Timeout() {
		t.Fatal("Timeout value is not consistent")
	}
	if config.reviveBurst != cfg.ReviveBurst() {
		t.Fatal("Revive burst value is not consistent")
	}
	if config.reviveWait != cfg.ReviveWait() {
		t.Fatal("Revive wait duration is not consistent")
	}
}

// Benchmarks setting up default configuration values
func BenchmarkConfiguration_Initialize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fs := flag.NewFlagSet("test", flag.PanicOnError)

		config := new(SprintConfiguration)
		config.Initialize(fs)
	}
}

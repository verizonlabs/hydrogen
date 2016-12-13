package scheduler

import (
	"flag"
	"github.com/verizonlabs/mesos-go"
	"testing"
	"time"
)

// Mocked configuration
type mockConfiguration struct {
	cfg Configuration
}

func (m *mockConfiguration) Initialize(fs *flag.FlagSet) {
	m.cfg.name = "Sprint"
	m.cfg.checkpointing = true
	m.cfg.command = ""
	m.cfg.endpoint = "http://127.0.0.1:5050/api/v1/scheduler"
	m.cfg.principal = "Sprint"
	m.cfg.reviveBurst = 3
	m.cfg.reviveWait = 1 * time.Second
	m.cfg.timeout = 20 * time.Second
}

func (m *mockConfiguration) GetName() string {
	return m.cfg.name
}

func (m *mockConfiguration) GetCheckpointing() *bool {
	return &m.cfg.checkpointing
}

func (m *mockConfiguration) GetPrincipal() string {
	return m.cfg.principal
}

func (m *mockConfiguration) GetCommand() string {
	return m.cfg.command
}

func (m *mockConfiguration) GetUris() []mesos.CommandInfo_URI {
	return m.cfg.uris
}

func (m *mockConfiguration) GetTimeout() time.Duration {
	return m.cfg.timeout
}

func (m *mockConfiguration) GetEndpoint() string {
	return m.cfg.endpoint
}

func (m *mockConfiguration) GetReviveBurst() int {
	return m.cfg.reviveBurst
}

func (m *mockConfiguration) GetReviveWait() time.Duration {
	return m.cfg.reviveWait
}

var cfg baseConfiguration

func init() {
	cfg = new(mockConfiguration)
}

// Tests setting up default configuration values
func TestConfiguration_Initialize(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)

	config := new(Configuration)
	config.Initialize(fs)

	if config.endpoint != cfg.GetEndpoint() {
		t.Fatal("Invalid endpoint")
	}
	if config.checkpointing != *cfg.GetCheckpointing() {
		t.Fatal("Checkpointing is disabled")
	}
	if config.command != cfg.GetCommand() {
		t.Fatal("Invalid command")
	}
	if config.name != cfg.GetName() {
		t.Fatal("Invalid framework name")
	}
	if config.principal != cfg.GetPrincipal() {
		t.Fatal("Invalid framework principal")
	}
	if config.timeout != cfg.GetTimeout() {
		t.Fatal("Timeout value is not consistent")
	}
	if config.reviveBurst != cfg.GetReviveBurst() {
		t.Fatal("Revive burst value is not consistent")
	}
	if config.reviveWait != cfg.GetReviveWait() {
		t.Fatal("Revive wait duration is not consistent")
	}
}

// Benchmarks setting up default configuration values
func BenchmarkConfiguration_Initialize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fs := flag.NewFlagSet("test", flag.PanicOnError)

		config := new(Configuration)
		config.Initialize(fs)
	}
}

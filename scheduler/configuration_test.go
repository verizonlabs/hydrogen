package scheduler

import (
	"flag"
	"testing"
	"time"
)

// Mocked configuration
type mockConfiguration struct {
	name          string
	checkpointing *bool
}

func (m *mockConfiguration) Initialize(fs *flag.FlagSet) {
	checkpointing := true

	m.name = "Test"
	m.checkpointing = &checkpointing
}

func (m *mockConfiguration) GetName() string {
	return m.name
}

func (m *mockConfiguration) GetCheckpointing() *bool {
	return m.checkpointing
}

var (
	endpoint      = "http://127.0.0.1:5050/api/v1/scheduler"
	name          = "Sprint"
	checkpointing = true
	principal     = "Sprint"
	command       = ""
	timeout       = 20 * time.Second
	reviveBurst   = 3
	reviveWait    = 1 * time.Second
)

// Tests setting up default configuration values
func TestConfiguration_Initialize(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)

	config := new(Configuration)
	config.Initialize(fs)

	if config.endpoint != endpoint {
		t.Fatal("Invalid endpoint")
	}
	if config.checkpointing != checkpointing {
		t.Fatal("Checkpointing is disabled")
	}
	if config.command != command {
		t.Fatal("Invalid command")
	}
	if config.name != name {
		t.Fatal("Invalid framework name")
	}
	if config.principal != principal {
		t.Fatal("Invalid framework principal")
	}
	if config.timeout != timeout {
		t.Fatal("Timeout value is not consistent")
	}
	if config.reviveBurst != reviveBurst {
		t.Fatal("Revive burst value is not consistent")
	}
	if config.reviveWait != reviveWait {
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

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
	m.cfg.maxRefuse = 5 * time.Second
	m.cfg.executorSrvPath = "executor"
	m.cfg.executorSrvPort = 8081

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

func (m *mockConfiguration) MaxRefuse() time.Duration {
	return m.cfg.maxRefuse
}

func (m *mockConfiguration) ExecutorSrvCert() string {
	return m.cfg.executorSrvCert
}

func (m *mockConfiguration) ExecutorSrvKey() string {
	return m.cfg.executorSrvKey
}

func (m *mockConfiguration) ExecutorSrvPath() string {
	return m.cfg.executorSrvPath
}

func (m *mockConfiguration) ExecutorSrvPort() int {
	return m.cfg.executorSrvPort
}

var cfg configuration = new(mockConfiguration).Initialize(nil)

// Tests setting up default configuration values
func TestSprintConfiguration_Initialize(t *testing.T) {
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
func BenchmarkSprintConfiguration_Initialize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fs := flag.NewFlagSet("test", flag.PanicOnError)

		config := new(SprintConfiguration)
		config.Initialize(fs)
	}
}

// Make sure we return the right name.
func TestSprintConfiguration_Name(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.Name() != "Sprint" {
		t.Fatal("Configuration has wrong name")
	}
}

// Checks to see if our default value for checkpointing is right.
func TestSprintConfiguration_Checkpointing(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if !*config.Checkpointing() {
		t.Fatal("Checkpointing is not set to the right value")
	}
}

// Make sure we have the right principal value.
func TestSprintConfiguration_Principal(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.Principal() != "Sprint" {
		t.Fatal("Principal is not set to the right value")
	}
}

// Checks to see whether the command is set properly.
func TestSprintConfiguration_Command(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if *config.Command() != "" {
		t.Fatal("Command is not set to the right value")
	}
}

// Make sure the URIs are set correctly.
func TestSprintConfiguration_Uris(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if len(config.Uris()) != 0 {
		t.Fatal("The number of URIs should be 0")
	}
}

// Check our default timeout value.
func TestSprintConfiguration_Timeout(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.Timeout() != 20*time.Second {
		t.Fatal("Timeout is not set to the right value")
	}
}

// Make sure we have the right default endpoint value.
func TestSprintConfiguration_Endpoint(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.Endpoint() != "http://127.0.0.1:5050/api/v1/scheduler" {
		t.Fatal("Endpoint is not set to the right value")
	}
}

// Ensure we have the right revive burst amount.
func TestSprintConfiguration_ReviveBurst(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.ReviveBurst() != 3 {
		t.Fatal("Revive burst is not set to the right value")
	}
}

// Ensure we have the right revive wait period.
func TestSprintConfiguration_ReviveWait(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.ReviveWait() != 1*time.Second {
		t.Fatal("Revive wait period is not set to the right value")
	}
}

// Ensure we have the right maximum refusal time.
func TestSprintConfiguration_MaxRefuse(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.MaxRefuse() != 5*time.Second {
		t.Fatal("Max refusal time is not set to the right value")
	}
}

// Make sure we get our TLS certificate properly.
func TestSprintConfiguration_ExecutorSrvCert(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.ExecutorSrvCert() != "tls.crt" {
		t.Fatal("TLS certificate is wrong")
	}
}

// Make sure we get our TLS key properly.
func TestSprintConfiguration_ExecutorSrvKey(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.ExecutorSrvKey() != "tls.key" {
		t.Fatal("TLS key is wrong")
	}
}

// Make sure we get our executor path properly.
func TestSprintConfiguration_ExecutorSrvPath(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.ExecutorSrvPath() != "executor" {
		t.Fatal("Executor binary path is wrong")
	}
}

// Make sure we get our executor port properly.
func TestSprintConfiguration_ExecutorSrvPort(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.PanicOnError)
	config := new(SprintConfiguration).Initialize(fs)
	if config.ExecutorSrvPort() != 8081 {
		t.Fatal("Executor server port is wrong")
	}
}

package scheduler

import (
	"mesos-sdk"
	"os/user"
	"sprint/scheduler/server"
	"testing"
	"time"
)

// Mocked configuration
type mockConfiguration struct {
	cfg SprintConfiguration
}

func (m *mockConfiguration) Initialize() *SprintConfiguration {
	m.cfg.name = "Sprint"
	m.cfg.user = "root"
	m.cfg.checkpointing = true
	m.cfg.command = ""
	m.cfg.endpoint = "http://127.0.0.1:5050/api/v1/scheduler"
	m.cfg.principal = "Sprint"
	m.cfg.reviveBurst = 3
	m.cfg.reviveWait = 1 * time.Second
	m.cfg.timeout = 20 * time.Second
	m.cfg.maxRefuse = 5 * time.Second
	m.cfg.executorSrvCfg = new(server.ServerConfiguration).Initialize()
	m.cfg.executorName = "Sprinter"
	m.cfg.executorCmd = "./executor"

	return &m.cfg
}

func (m *mockConfiguration) Name() string {
	return m.cfg.name
}

func (m *mockConfiguration) User() string {
	return m.cfg.user
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

func (m *mockConfiguration) SetExecutorSrvCfg(cfg server.Configuration) *SprintConfiguration {
	m.cfg.executorSrvCfg = cfg

	return &m.cfg
}

func (m *mockConfiguration) ExecutorSrvCfg() server.Configuration {
	return m.cfg.executorSrvCfg
}

func (m *mockConfiguration) ExecutorName() *string {
	return &m.cfg.executorName
}

func (m *mockConfiguration) ExecutorCmd() *string {
	return &m.cfg.executorCmd
}

var cfg configuration = new(mockConfiguration).Initialize()
var sprintConfig = new(SprintConfiguration).Initialize()

// Tests setting up default configuration values
func TestSprintConfiguration_Initialize(t *testing.T) {
	t.Parallel()

	if sprintConfig.endpoint != cfg.Endpoint() {
		t.Fatal("Invalid endpoint")
	}
	if sprintConfig.checkpointing != *cfg.Checkpointing() {
		t.Fatal("Checkpointing is disabled")
	}
	if sprintConfig.command != *cfg.Command() {
		t.Fatal("Invalid command")
	}
	if sprintConfig.name != cfg.Name() {
		t.Fatal("Invalid framework name")
	}
	if sprintConfig.principal != cfg.Principal() {
		t.Fatal("Invalid framework principal")
	}
	if sprintConfig.timeout != cfg.Timeout() {
		t.Fatal("Timeout value is not consistent")
	}
	if sprintConfig.reviveBurst != cfg.ReviveBurst() {
		t.Fatal("Revive burst value is not consistent")
	}
	if sprintConfig.reviveWait != cfg.ReviveWait() {
		t.Fatal("Revive wait duration is not consistent")
	}
}

// Make sure we return the right name.
func TestSprintConfiguration_Name(t *testing.T) {
	t.Parallel()

	if sprintConfig.Name() != "Sprint" {
		t.Fatal("Configuration has wrong name")
	}
}

// Ensures that we can detect the current user and pass it into the framework info.
func TestSprintConfiguration_User(t *testing.T) {
	t.Parallel()

	u, err := user.Current()
	if err != nil {
		t.Fatal("Unable to detect current user: " + err.Error())
	}

	if sprintConfig.User() != u.Username {
		t.Fatal("User is not set correctly")
	}
}

// Checks to see if our default value for checkpointing is right.
func TestSprintConfiguration_Checkpointing(t *testing.T) {
	t.Parallel()

	if !*sprintConfig.Checkpointing() {
		t.Fatal("Checkpointing is not set to the right value")
	}
}

// Make sure we have the right principal value.
func TestSprintConfiguration_Principal(t *testing.T) {
	t.Parallel()

	if sprintConfig.Principal() != "Sprint" {
		t.Fatal("Principal is not set to the right value")
	}
}

// Checks to see whether the command is set properly.
func TestSprintConfiguration_Command(t *testing.T) {
	t.Parallel()

	if *sprintConfig.Command() != "" {
		t.Fatal("Command is not set to the right value")
	}
}

// Make sure the URIs are set correctly.
func TestSprintConfiguration_Uris(t *testing.T) {
	t.Parallel()

	if len(sprintConfig.Uris()) != 0 {
		t.Fatal("The number of URIs should be 0")
	}
}

// Check our default timeout value.
func TestSprintConfiguration_Timeout(t *testing.T) {
	t.Parallel()

	if sprintConfig.Timeout() != 20*time.Second {
		t.Fatal("Timeout is not set to the right value")
	}
}

// Make sure we have the right default endpoint value.
func TestSprintConfiguration_Endpoint(t *testing.T) {
	t.Parallel()

	if sprintConfig.Endpoint() != "http://127.0.0.1:5050/api/v1/scheduler" {
		t.Fatal("Endpoint is not set to the right value")
	}
}

// Ensure we have the right revive burst amount.
func TestSprintConfiguration_ReviveBurst(t *testing.T) {
	t.Parallel()

	if sprintConfig.ReviveBurst() != 3 {
		t.Fatal("Revive burst is not set to the right value")
	}
}

// Ensure we have the right revive wait period.
func TestSprintConfiguration_ReviveWait(t *testing.T) {
	t.Parallel()

	if sprintConfig.ReviveWait() != 1*time.Second {
		t.Fatal("Revive wait period is not set to the right value")
	}
}

// Ensure we have the right maximum refusal time.
func TestSprintConfiguration_MaxRefuse(t *testing.T) {
	t.Parallel()

	if sprintConfig.MaxRefuse() != 5*time.Second {
		t.Fatal("Max refusal time is not set to the right value")
	}
}

// Make sure we get our executor name correctly.
func TestSprintConfiguration_ExecutorName(t *testing.T) {
	t.Parallel()

	if *sprintConfig.ExecutorName() != "Sprinter" {
		t.Fatal("Executor name is wrong")
	}
}

// Make sure we get our executor command correctly.
func TestSprintConfiguration_ExecutorCmd(t *testing.T) {
	t.Parallel()

	if *sprintConfig.ExecutorCmd() != "./executor" {
		t.Fatal("Executor command is wrong")
	}
}

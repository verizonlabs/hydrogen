package server

import (
	"testing"
)

// Mocked configuration
type mockConfiguration struct {
	cfg ServerConfiguration
}

func (m *mockConfiguration) Initialize() *ServerConfiguration {
	m.cfg.executorSrvPath = "executor"
	m.cfg.executorSrvPort = 8081

	return &m.cfg
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

var cfg Configuration = new(mockConfiguration).Initialize()
var serverCfg = new(ServerConfiguration).Initialize()

// Make sure we get our TLS certificate properly.
func TestServerConfiguration_ExecutorSrvCert(t *testing.T) {
	t.Parallel()

	if serverCfg.ExecutorSrvCert() != "" {
		t.Fatal("TLS certificate is wrong")
	}
}

// Make sure we get our TLS key properly.
func TestServerConfiguration_ExecutorSrvKey(t *testing.T) {
	t.Parallel()

	if serverCfg.ExecutorSrvKey() != "" {
		t.Fatal("TLS key is wrong")
	}
}

// Make sure we get our executor path properly.
func TestServerConfiguration_ExecutorSrvPath(t *testing.T) {
	t.Parallel()

	if serverCfg.ExecutorSrvPath() != "executor" {
		t.Fatal("Executor binary path is wrong")
	}
}

// Make sure we get our executor port properly.
func TestServerConfiguration_ExecutorSrvPort(t *testing.T) {
	t.Parallel()

	if serverCfg.ExecutorSrvPort() != 8081 {
		t.Fatal("Executor server port is wrong")
	}
}

// Make sure our protocol is set correctly.
func TestServerConfiguration_ExecutorSrvProtocol(t *testing.T) {
	t.Parallel()

	if serverCfg.ExecutorSrvProtocol() != "http" {
		t.Fatal("Executor server protocol is incorrect")
	}
}

package file

import (
	"reflect"
	"sprint/scheduler/server"
	"testing"
)

// Mocked configuration
type mockConfiguration struct {
	cfg server.ServerConfiguration
}

func (m *mockConfiguration) Initialize() *server.ServerConfiguration {
	*m.cfg.ExecutorSrvPath() = "executor"
	*m.cfg.ExecutorSrvPort() = 8081

	return &m.cfg
}

func (m *mockConfiguration) ExecutorSrvCert() string {
	return m.cfg.ExecutorSrvCert()
}

func (m *mockConfiguration) ExecutorSrvKey() string {
	return m.cfg.ExecutorSrvKey()
}

func (m *mockConfiguration) ExecutorSrvPath() *string {
	return m.cfg.ExecutorSrvPath()
}

func (m *mockConfiguration) ExecutorSrvPort() *int {
	return m.cfg.ExecutorSrvPort()
}

var cfg server.Configuration = new(mockConfiguration).Initialize()

// Make sure we get the right type for our executor server.
func TestNewExecutorServer(t *testing.T) {
	t.Parallel()

	path := "executor"
	port := 8081
	cert := ""
	key := ""

	srv := NewExecutorServer(cfg)
	if reflect.TypeOf(srv) != reflect.TypeOf(new(executorServer)) {
		t.Fatal("Executor server is of the wrong type")
	}

	if srv.path != path {
		t.Fatal("Executor server path was not set correctly")
	}
	if srv.port != port {
		t.Fatal("Executor server port was not set correctly")
	}
	if srv.cert != cert {
		t.Fatal("Executor server certificate was not set correctly")
	}
	if srv.key != key {
		t.Fatal("Executor server key was not set correctly")
	}
}

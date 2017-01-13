package server

import (
	"reflect"
	"testing"
)

// Make sure we get the right type for our executor server.
func TestNewExecutorServer(t *testing.T) {
	t.Parallel()

	path := "executor"
	port := 8081
	cert := ""
	key := ""

	server := NewExecutorServer(cfg)
	if reflect.TypeOf(server) != reflect.TypeOf(new(executorServer)) {
		t.Fatal("Executor server is of the wrong type")
	}

	if server.path != path {
		t.Fatal("Executor server path was not set correctly")
	}
	if server.port != port {
		t.Fatal("Executor server port was not set correctly")
	}
	if server.cert != cert {
		t.Fatal("Executor server certificate was not set correctly")
	}
	if server.key != key {
		t.Fatal("Executor server key was not set correctly")
	}
}

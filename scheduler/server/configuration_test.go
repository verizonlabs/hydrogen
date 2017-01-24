package server

import (
	"testing"
)

var serverCfg = new(ServerConfiguration).Initialize()

// Make sure we get our TLS certificate properly.
func TestServerConfiguration_ExecutorSrvCert(t *testing.T) {
	t.Parallel()

	if serverCfg.Cert() != "" {
		t.Fatal("TLS certificate is wrong")
	}
}

// Make sure we get our TLS key properly.
func TestServerConfiguration_ExecutorSrvKey(t *testing.T) {
	t.Parallel()

	if serverCfg.Key() != "" {
		t.Fatal("TLS key is wrong")
	}
}

// Make sure we get our executor port properly.
func TestServerConfiguration_ExecutorSrvPort(t *testing.T) {
	t.Parallel()

	if *serverCfg.Port() != 8081 {
		t.Fatal("Executor server port is wrong")
	}
}

// Make sure our protocol is set correctly.
func TestServerConfiguration_ExecutorSrvProtocol(t *testing.T) {
	t.Parallel()

	if serverCfg.Protocol() != "http" {
		t.Fatal("Executor server protocol is incorrect")
	}
}

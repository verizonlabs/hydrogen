package server

import (
	"testing"
)

var serverCfg = new(ServerConfiguration).Initialize()

// Make sure we get our TLS certificate properly.
func TestServerConfiguration_Cert(t *testing.T) {
	t.Parallel()

	if serverCfg.Cert() != "" {
		t.Fatal("TLS certificate is wrong")
	}
}

// Make sure we get our TLS key properly.
func TestServerConfiguration_Key(t *testing.T) {
	t.Parallel()

	if serverCfg.Key() != "" {
		t.Fatal("TLS key is wrong")
	}
}

// Make sure our protocol is set correctly.
func TestServerConfiguration_Protocol(t *testing.T) {
	t.Parallel()

	if serverCfg.Protocol() != "http" {
		t.Fatal("Server protocol is incorrect")
	}
}

// Make sure TLS defaults to off.
func TestServerConfiguration_TLS(t *testing.T) {
	t.Parallel()

	if serverCfg.TLS() {
		t.Fatal("TLS has the wrong default setting")
	}
}

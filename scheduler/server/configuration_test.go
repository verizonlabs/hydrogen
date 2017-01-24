package server

import (
	"crypto/tls"
	"net/http"
	"reflect"
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

// Checks to see if our HTTP server is configured properly.
func TestServerConfiguration_Server(t *testing.T) {
	t.Parallel()

	srv := &http.Server{
		TLSConfig: &tls.Config{
			// Use only the most secure protocol version.
			MinVersion: tls.VersionTLS12,
			// Use very strong crypto curves.
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			// Use very strong cipher suites (order is important here!)
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Required for HTTP/2 support.
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
	}

	if !reflect.DeepEqual(serverCfg.Server(), srv) {
		t.Fatal("HTTP server is not initialized correctly")
	}
}

// Make sure TLS defaults to off.
func TestServerConfiguration_TLS(t *testing.T) {
	t.Parallel()

	if serverCfg.TLS() {
		t.Fatal("TLS has the wrong default setting")
	}
}

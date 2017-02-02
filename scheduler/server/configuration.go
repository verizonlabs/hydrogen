package server

import (
	"crypto/tls"
	"flag"
	"net/http"
)

type Configuration interface {
	Initialize() *ServerConfiguration
	IpAddress() string
	Cert() string
	Key() string
	Protocol() string
	Server() *http.Server
	TLS() bool
	GetEndpoint() string
}

// Configuration for the executor server.
type ServerConfiguration struct {
	ip     string
	cert   string
	key    string
	path   string
	server *http.Server
	tls    bool
}

func (c *ServerConfiguration) GetEndpoint() string {
	port := flag.Lookup("server.executor.port").Value.String()
	return c.Protocol() + "://" + c.ip + ":" + port
}

// Applies values to the various configurations from user-supplied flags.
func (c *ServerConfiguration) Initialize() *ServerConfiguration {
	flag.StringVar(&c.cert, "server.cert", "", "TLS certificate")
	flag.StringVar(&c.key, "server.key", "", "TLS key")
	flag.StringVar(&c.ip, "server.ip", "10.0.2.2", "IP address for the server to listen on.")

	c.server = &http.Server{
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

	return c
}

// Gets the path to the TLS certificate.
func (c *ServerConfiguration) Cert() string {
	return c.cert
}

// Gets the path to the TLS key.
func (c *ServerConfiguration) Key() string {
	return c.key
}

// Determines the protocol to be used.
func (c *ServerConfiguration) Protocol() string {
	if c.tls {
		return "https"
	} else {
		return "http"
	}
}

// Returns the custom HTTP server with TLS configuration.
func (c *ServerConfiguration) Server() *http.Server {
	return c.server
}

// If a TLS certificate and key have been provided then TLS is enabled.
func (c *ServerConfiguration) TLS() bool {
	return c.cert != "" && c.key != ""
}

// Return the IP address as a string.
func (c *ServerConfiguration) IpAddress() string {
	return c.ip
}

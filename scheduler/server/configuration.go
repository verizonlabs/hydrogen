package server

import (
	"flag"
)

type Configuration interface {
	Initialize() *ServerConfiguration
	Cert() string
	Key() string
	Protocol() string
	Port() int
}

// Configuration for the executor server.
type ServerConfiguration struct {
	cert string
	key  string
	path string
	port int
}

// Applies values to the various configurations from user-supplied flags.
func (c *ServerConfiguration) Initialize() *ServerConfiguration {
	flag.StringVar(&c.cert, "server.cert", "", "TLS certificate")
	flag.StringVar(&c.key, "server.key", "", "TLS key")

	return c
}

func (c *ServerConfiguration) Cert() string {
	return c.cert
}

func (c *ServerConfiguration) Key() string {
	return c.key
}

func (c *ServerConfiguration) Protocol() string {
	if c.cert != "" && c.key != "" {
		return "https"
	} else {
		return "http"
	}
}

func (c *ServerConfiguration) Port() int {
	return c.port
}

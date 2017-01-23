package server

import (
	"flag"
)

type Configuration interface {
	Initialize() *ServerConfiguration
	Cert() string
	Key() string
	Path() *string
	Port() *int
	Protocol() string
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
	flag.StringVar(&c.cert, "server.executor.cert", "", "TLS certificate")
	flag.StringVar(&c.key, "server.executor.key", "", "TLS key")
	flag.StringVar(&c.path, "server.executor.path", "executor", "Path to the executor binary")
	flag.IntVar(&c.port, "server.executor.port", 8081, "Executor server listen port")

	return c
}

func (c *ServerConfiguration) Cert() string {
	return c.cert
}

func (c *ServerConfiguration) Key() string {
	return c.key
}

func (c *ServerConfiguration) Path() *string {
	return &c.path
}

func (c *ServerConfiguration) Port() *int {
	return &c.port
}

func (c *ServerConfiguration) Protocol() string {
	if c.cert != "" && c.key != "" {
		return "https"
	} else {
		return "http"
	}
}

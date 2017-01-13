package server

import (
	"flag"
)

type Configuration interface {
	Initialize() *ServerConfiguration
	ExecutorSrvCert() string
	ExecutorSrvKey() string
	ExecutorSrvPath() string
	ExecutorSrvPort() int
}

// Configuration for the executor server.
type ServerConfiguration struct {
	executorSrvCert string
	executorSrvKey  string
	executorSrvPath string
	executorSrvPort int
}

// Applies values to the various configurations from user-supplied flags.
func (c *ServerConfiguration) Initialize() *ServerConfiguration {
	flag.StringVar(&c.executorSrvCert, "server.executor.cert", "", "TLS certificate")
	flag.StringVar(&c.executorSrvKey, "server.executor.key", "", "TLS key")
	flag.StringVar(&c.executorSrvPath, "server.executor.path", "executor", "Path to the executor binary")
	flag.IntVar(&c.executorSrvPort, "server.executor.port", 8081, "Executor server listen port")

	return c
}

func (c *ServerConfiguration) ExecutorSrvCert() string {
	return c.executorSrvCert
}

func (c *ServerConfiguration) ExecutorSrvKey() string {
	return c.executorSrvKey
}

func (c *ServerConfiguration) ExecutorSrvPath() string {
	return c.executorSrvPath
}

func (c *ServerConfiguration) ExecutorSrvPort() int {
	return c.executorSrvPort
}

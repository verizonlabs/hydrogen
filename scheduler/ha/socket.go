package ha

import (
	"mesos-framework-sdk/logging"
	"net"
	"sprint/scheduler"
	"strconv"
	"time"
)

// Handles connections from other framework instances that try and determine the state of the leader.
// Used in coordination with determining if and when we need to perform leader election.
func LeaderServer(c *scheduler.Configuration, logger logging.Logger) {
	addr, err := net.ResolveTCPAddr(c.Leader.AddressFamily, "["+c.Leader.IP+"]:"+strconv.Itoa(c.Leader.ServerPort))
	if err != nil {
		logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	tcp, err := net.ListenTCP(c.Leader.AddressFamily, addr)
	if err != nil {
		logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	for {

		// Block here until we get a new connection.
		// We don't want to do anything with the stream so move on without spawning a thread to handle the connection.
		conn, err := tcp.AcceptTCP()
		if err != nil {
			logger.Emit(logging.ERROR, "Failed to accept client: %s", err.Error())
			time.Sleep(c.Leader.ServerRetry)
			continue
		}

		// TODO build out some config to use for setting the keep alive period here
		if err := conn.SetKeepAlive(true); err != nil {
			logger.Emit(logging.ERROR, "Failed to set keep alive: %s", err.Error())
		}
	}
}

// Connects to the leader and determines if and when we should start the leader election process.
func LeaderClient(c *scheduler.Configuration, leader string) error {
	conn, err := net.DialTimeout(c.Leader.AddressFamily, "["+leader+"]:"+strconv.Itoa(c.Leader.ServerPort), 2*time.Second) // TODO make this configurable?
	if err != nil {
		return err
	}

	// TODO build out some config to use for setting the keep alive period here
	tcp := conn.(*net.TCPConn)
	if err := tcp.SetKeepAlive(true); err != nil {
		return err
	}

	buffer := make([]byte, 1)
	for {
		_, err := tcp.Read(buffer)
		if err != nil {
			return err
		}
	}
}

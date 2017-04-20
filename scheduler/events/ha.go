package events

import (
	"mesos-framework-sdk/ha"
	"mesos-framework-sdk/logging"
	"net"
	"strconv"
	"time"
)

//
// High Availability (HA) handles satisfying the interface for HA as defined by the SDK.
//

// Getters
func (s *SprintEventController) Name() (string, error) {
	return s.name, nil
}

func (s *SprintEventController) Status() (ha.Status, error) {
	return s.status, nil
}

//
// Communicate is a public method that defines how we communicate in our HA implementation.
// Here we set up a listening mechanism on a TCP channel for other nodes to talk to us.
//
func (s *SprintEventController) Communicate() {
	// NOTE (tim): Using a generic Listener might be better so we don't have to assume TCP
	// for all implementations in future?
	addr, err := net.ResolveTCPAddr(s.config.Leader.AddressFamily, "["+s.config.Leader.IP+"]:"+
		strconv.Itoa(s.config.Leader.ServerPort))
	if err != nil {
		s.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	// We start listening for connections.
	tcp, err := net.ListenTCP(s.config.Leader.AddressFamily, addr)
	if err != nil {
		s.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	for {
		// Block here until we get a new connection.
		// We don't want to do anything with the stream so move on without spawning a thread to handle the connection.
		conn, err := tcp.AcceptTCP()
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to accept client: %s", err.Error())
			time.Sleep(s.config.Leader.ServerRetry)
			continue
		}

		// TODO build out some config to use for setting the keep alive period here
		if err := conn.SetKeepAlive(true); err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set keep alive: %s", err.Error())
		}
	}

}

//
// Election defines how we elect our leader in our HA mode.
// We attempt to write ourselves as the leader in etcd.
// If we cannot write and their is another leader already up, we simply
// become a client and send a single byte via the TCP channel as a heartbeat.
// Should the leader ever go down, we start the election again.
// This is a simple approach in which who ever gets to write to etcd first
// becomes the leader.
//
func (s *SprintEventController) Election() {
	for {
		// This will only set us as the leader if there isn't an already existing leader.
		s.CreateLeader()

		leader := s.GetLeader()
		if leader != s.config.Leader.IP {
			s.status = ha.Listening
			s.logger.Emit(logging.INFO, "Connecting to leader to determine when we need to wake up and perform leader election")

			// Block here until we lose connection to the leader.
			// Once the connection has been lost elect a new leader.
			err := s.leaderClient(leader)

			// Only delete the key if we've lost the connection, not timed out.
			// This conditional requires Go 1.6+

			// NOTE (tim): Casting here is dangerous and could lead to a panic
			if err, ok := err.(net.Error); ok && err.Timeout() {
				s.logger.Emit(logging.ERROR, "Timed out connecting to leader")
			} else {
				s.status = ha.Election
				s.logger.Emit(logging.ERROR, "Lost connection to leader")
				s.kv.Delete("/leader")
			}
		} else {
			s.status = ha.Leading
			s.logger.Emit(logging.INFO, "We're leading.")
			// We are the leader, exit the loop and start the scheduler/API.
			break
		}
	}
}

//
// Connects to the leader and determines if and when we should start the leader election process.
//
func (s *SprintEventController) leaderClient(leader string) error {
	conn, err := net.DialTimeout(s.config.Leader.AddressFamily, "["+leader+"]:"+
		strconv.Itoa(s.config.Leader.ServerPort), 2*time.Second)
	if err != nil {
		return err
	}

	s.logger.Emit(logging.INFO, "Successfully connected to leader %s", leader)

	// NOTE (tim): This cast has the potential to panic.
	tcp := conn.(*net.TCPConn)
	if err := tcp.SetKeepAlive(true); err != nil {
		return err
	}

	s.status = ha.Talking
	buffer := make([]byte, 1)
	for {
		_, err := tcp.Read(buffer)
		if err != nil {
			return err
		}
	}
}

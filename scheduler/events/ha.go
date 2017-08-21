// Copyright 2017 Verizon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"mesos-framework-sdk/ha"
	"mesos-framework-sdk/logging"
	"net"
	"os"
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
	// for all implementations in future.
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
// All running schedulers attempt to write themselves as the leader to etcd.
// If there is another existing leader we connect to the leader as a client using TCP keepalive.
// Should the leader ever go down, we start the election again and wait for an instance to win the election.
// This is a simple approach in which who ever gets to write to etcd first becomes the leader.
// Note that reads and writes are atomic operations.
//
func (s *SprintEventController) Election() {
	for {
		// This will only set us as the leader if there isn't an already existing leader.
		// If there's an already existing leader then this call is effectively a no-op.
		err := s.CreateLeader()
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to persist leader information: %s", err.Error())
			os.Exit(3)
		}

		leader, err := s.GetLeader()
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get leader information: %s", err.Error())
			os.Exit(5)
		}

		// If the leader fetched from persistent storage does not match the IP provided to this instance
		// then we should connect to the already existing leader.
		if leader != s.config.Leader.IP {
			s.status = ha.Listening
			s.logger.Emit(logging.INFO, "Connecting to leader to determine when we need to wake up and perform leader election")

			// Block here until we lose connection to the leader.
			// Once the connection has been lost, elect a new leader.
			err := s.leaderClient(leader)

			// Only delete the key if we've lost the connection, not timed out.
			// This conditional requires Go 1.6+
			// NOTE (tim): Casting here is dangerous and could lead to a panic
			if err, ok := err.(net.Error); ok && err.Timeout() {
				s.logger.Emit(logging.ERROR, "Timed out connecting to leader")
			} else {
				s.status = ha.Election
				s.logger.Emit(logging.ERROR, "Lost connection to leader")
				err := s.deleteLeader()
				if err != nil {
					s.logger.Emit(logging.ERROR, "Failed to delete leader information: %s", err.Error())
					os.Exit(4)
				}
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

	// Currently the leader server does not send us anything as we rely on TCP keepalive.
	// Allocate the smallest buffer we can and block on reading data.
	// If we don't do this we'll spin like crazy in an empty loop.
	buffer := make([]byte, 1)
	for {
		_, err := tcp.Read(buffer)
		if err != nil {
			return err
		}
	}
}

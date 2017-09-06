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

package ha

import (
	"mesos-framework-sdk/logging"
	"net"
	"os"
	"hydrogen/scheduler"
	"hydrogen/task/persistence"
	"strconv"
	"time"
)

const (
	leaderKey = "/leader"
)

type HA struct {
	logger  logging.Logger
	config  *scheduler.LeaderConfiguration
	storage persistence.Storage
}

func NewHA(s persistence.Storage, l logging.Logger, c *scheduler.LeaderConfiguration) *HA {
	return &HA{
		logger:  l,
		config:  c,
		storage: s,
	}
}

// Communicate is a public method that defines how we communicate in our HA implementation.
// Here we set up a listening mechanism on a TCP channel for other nodes to talk to us.
func (h *HA) Communicate() {
	// NOTE (tim): Using a generic Listener might be better so we don't have to assume TCP
	// for all implementations in future.
	addr, err := net.ResolveTCPAddr(h.config.AddressFamily, "["+h.config.IP+"]:"+
		strconv.Itoa(h.config.ServerPort))
	if err != nil {
		h.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	// We start listening for connections.
	tcp, err := net.ListenTCP(h.config.AddressFamily, addr)
	if err != nil {
		h.logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	for {

		// Block here until we get a new connection.
		conn, err := tcp.AcceptTCP()
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to accept client: %s", err.Error())
			time.Sleep(h.config.ServerRetry)
			continue
		}

		remote := conn.RemoteAddr().String()
		host, _, err := net.SplitHostPort(remote)
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to parse IP and port of incoming standby connection")
			return
		}

		h.logger.Emit(logging.INFO, "Standby "+host+" connected")
		go func(conn *net.TCPConn) {
			buff := make([]byte, 1)
			_, err := conn.Read(buff)

			// This cast requires Go 1.6+
			if err, ok := err.(net.Error); ok && err.Timeout() {
				h.logger.Emit(logging.ERROR, "Connection from "+host+" timed out")
			} else {
				h.logger.Emit(logging.ERROR, "Lost connection to standby "+host)
			}
		}(conn)

		// TODO build out some config to use for setting the keep alive period here
		if err := conn.SetKeepAlive(true); err != nil {
			h.logger.Emit(logging.ERROR, "Failed to set keep alive: %s", err.Error())
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
func (h *HA) Election() {
	for {
		// This will only set us as the leader if there isn't an already existing leader.
		// If there's an already existing leader then this call is effectively a no-op.
		err := h.CreateLeader()
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to persist leader information: %s", err.Error())
			os.Exit(3)
		}

		leader, err := h.GetLeader()
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to get leader information: %s", err.Error())
			os.Exit(5)
		}

		// If the leader fetched from persistent storage does not match the IP provided to this instance
		// then we should connect to the already existing leader.
		if leader != h.config.IP {
			h.logger.Emit(logging.INFO, "Connecting to existing leader")

			// Block here until we lose connection to the leader.
			// Once the connection has been lost, elect a new leader.
			err := h.leaderClient(leader)

			// Only delete the key if we've lost the connection, not timed out.
			// This cast requires Go 1.6+
			// NOTE (tim): Casting here is dangerous and could lead to a panic
			if err, ok := err.(net.Error); ok && err.Timeout() {
				h.logger.Emit(logging.ERROR, "Timed out connecting to leader")
			} else {
				h.logger.Emit(logging.ERROR, "Lost connection to leader")
				err := h.deleteLeader()
				if err != nil {
					h.logger.Emit(logging.ERROR, "Failed to delete leader information: %s", err.Error())
					os.Exit(4)
				}
			}
		} else {
			h.logger.Emit(logging.INFO, "We're leading")
			break // We are the leader, exit the loop and start the scheduler/API.
		}
	}
}

//
// Connects to the leader and determines if and when we should start the leader election process.
//
func (h *HA) leaderClient(leader string) error {
	conn, err := net.DialTimeout(h.config.AddressFamily, "["+leader+"]:"+
		strconv.Itoa(h.config.ServerPort), 2*time.Second)
	if err != nil {
		return err
	}

	h.logger.Emit(logging.INFO, "Successfully connected to leader %s", leader)

	// NOTE (tim): This cast has the potential to panic.
	tcp := conn.(*net.TCPConn)
	if err := tcp.SetKeepAlive(true); err != nil {
		return err
	}

	// Currently the leader server does not send us anything as we rely on TCP keepalive.
	// Allocate the smallest buffer we can and block on reading data.
	buffer := make([]byte, 1)
	_, err = tcp.Read(buffer)

	return err
}

// Deletes the current leader information.
func (h *HA) deleteLeader() error {
	policy := h.storage.CheckPolicy(nil)
	return h.storage.RunPolicy(policy, func() error {
		err := h.storage.Delete(leaderKey)
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to delete leader: %s", err.Error())
		}

		return err
	})
}

// Atomically create leader information.
func (h *HA) CreateLeader() error {
	policy := h.storage.CheckPolicy(nil)
	return h.storage.RunPolicy(policy, func() error {
		err := h.storage.Create(leaderKey, h.config.IP)
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to set leader: %s", err.Error())
		}

		return err
	})
}

// Atomically get leader information.
func (h *HA) GetLeader() (string, error) {
	var leader string
	policy := h.storage.CheckPolicy(nil)
	err := h.storage.RunPolicy(policy, func() error {
		l, err := h.storage.Read(leaderKey)
		if err != nil {
			h.logger.Emit(logging.ERROR, "Failed to get the leader: %s", err.Error())
			return err
		}

		leader = l
		return nil
	})

	if err != nil {
		return "", err
	}

	return leader, nil
}

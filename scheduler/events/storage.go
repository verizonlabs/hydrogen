package events

import (
	"mesos-framework-sdk/logging"
	"time"
)

//
// Storage is where the sprint event controller handles writing to etcd or any other storage backend.
//

// Atomically create leader information.
func (s *SprintEventController) CreateLeader() {
	for {
		if err := s.kv.Create("/leader", s.config.Leader.IP); err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set leader information: "+err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		break
	}

}

// Atomically get leader information.
func (s *SprintEventController) GetLeader() string {
	for {
		leader, err := s.kv.Read("/leader")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get the leader: %s", err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}

		return leader
	}
}

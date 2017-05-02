package events

import (
	"mesos-framework-sdk/include/mesos_v1"
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

func (s *SprintEventController) setFrameworkId() {
	for {
		id, err := s.kv.Read("/frameworkId")
		if err == nil {
			s.scheduler.FrameworkInfo().Id = &mesos_v1.FrameworkID{Value: &id}
			return
		} else {
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
	}
}

func (s *SprintEventController) readLeader() string {
	for {
		leader, err := s.kv.Read("/leader")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to find the leader: %s", err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		return leader
	}
}

func (s *SprintEventController) deleteLeader() {
	s.kv.Delete("/leader")
}

func (s *SprintEventController) createLeaderLease(idVal string) {
	for {
		lease, err := s.kv.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.FrameworkInfo().GetFailoverTimeout()))
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}

		s.lock.Lock()
		s.frameworkLease = lease
		s.lock.Unlock()

		return
	}
}
func (s *SprintEventController) refreshLeaderLease() {
	for {
		s.lock.RLock()
		if err := s.kv.RefreshLease(s.frameworkLease); err != nil {
			s.logger.Emit(logging.ERROR, "Failed to refresh framework ID lease: %s", err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		s.lock.RUnlock()

		return
	}
}

func (s *SprintEventController) getAllTasks() map[string]string {
	for {
		tasks, err := s.kv.ReadAll("/tasks")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get all task data: %s", err.Error())
			time.Sleep(s.config.Persistence.RetryInterval)
			continue
		}
		return tasks
	}
}

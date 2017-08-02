package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
)

//
// Storage is where the sprint event controller handles writing to etcd or any other storage backend.
//

// Atomically create leader information.
func (s *SprintEventController) CreateLeader() error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		err := s.storage.Create("/leader", s.config.Leader.IP)
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set leader: %s", err.Error())
		}

		return err
	})
}

// Atomically get leader information.
func (s *SprintEventController) GetLeader() (string, error) {
	var leader string
	policy := s.storage.CheckPolicy(nil)
	err := s.storage.RunPolicy(policy, func() error {
		l, err := s.storage.Read("/leader")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get the leader: %s", err.Error())
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

// Set our framework ID in memory from what's currently persisted.
// The framework ID is needed by the scheduler for almost any call to Mesos.
func (s *SprintEventController) setFrameworkId() error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		id, err := s.storage.Read("/frameworkId")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set the framework ID: %s", err.Error())
			return err
		}

		s.scheduler.FrameworkInfo().Id = &mesos_v1.FrameworkID{Value: &id}
		return nil
	})
}

// Deletes the current leader information.
func (s *SprintEventController) deleteLeader() error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		err := s.storage.Delete("/leader")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to delete leader: %s", err.Error())
		}

		return err
	})
}

// Create and persist our framework ID with an attached lifetime.
func (s *SprintEventController) createFrameworkIdLease(idVal string) error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		lease, err := s.storage.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.FrameworkInfo().GetFailoverTimeout()))
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to save framework ID of %s to persistent data store", idVal)
			return err
		}

		s.lock.Lock()
		s.frameworkLease = lease
		s.lock.Unlock()

		return nil
	})
}

// Refreshes the lifetime of our persisted framework ID.
func (s *SprintEventController) refreshFrameworkIdLease() error {
	policy := s.storage.CheckPolicy(nil)
	return s.storage.RunPolicy(policy, func() error {
		s.lock.RLock()
		err := s.storage.RefreshLease(s.frameworkLease)
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to refresh framework ID lease: %s", err.Error())
		}

		s.lock.RUnlock()

		return err
	})
}
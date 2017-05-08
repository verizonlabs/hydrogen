package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"sprint/task/retry"
)

//
// Storage is where the sprint event controller handles writing to etcd or any other storage backend.
//

// Atomically create leader information.
func (s *SprintEventController) CreateLeader() error {
	return s.taskmanager.RunPolicy(&retry.TaskRetry{
		MaxRetries: s.config.Persistence.MaxRetries,
		Backoff:    true,
	}, func() error {
		err := s.kv.Create("/leader", s.config.Leader.IP)
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set leader: %s", err.Error())
		}

		return err
	})
}

// Atomically get leader information.
func (s *SprintEventController) GetLeader() (string, error) {
	var leader string
	err := s.taskmanager.RunPolicy(&retry.TaskRetry{
		MaxRetries: s.config.Persistence.MaxRetries,
		Backoff:    true,
	}, func() error {
		l, err := s.kv.Read("/leader")
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

func (s *SprintEventController) setFrameworkId() error {
	return s.taskmanager.RunPolicy(&retry.TaskRetry{
		MaxRetries: s.config.Persistence.MaxRetries,
		Backoff:    true,
	}, func() error {
		id, err := s.kv.Read("/frameworkId")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to set the framework ID: %s", err.Error())
			return err
		}

		s.scheduler.FrameworkInfo().Id = &mesos_v1.FrameworkID{Value: &id}
		return nil
	})
}

func (s *SprintEventController) deleteLeader() error {
	return s.taskmanager.RunPolicy(&retry.TaskRetry{
		MaxRetries: s.config.Persistence.MaxRetries,
		Backoff:    true,
	}, func() error {
		err := s.kv.Delete("/leader")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to delete leader: %s", err.Error())
		}

		return err
	})

}

func (s *SprintEventController) createLeaderLease(idVal string) error {
	return s.taskmanager.RunPolicy(&retry.TaskRetry{
		MaxRetries: s.config.Persistence.MaxRetries,
		Backoff:    true,
	}, func() error {
		lease, err := s.kv.CreateWithLease("/frameworkId", idVal, int64(s.scheduler.FrameworkInfo().GetFailoverTimeout()))
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

func (s *SprintEventController) refreshLeaderLease() error {
	return s.taskmanager.RunPolicy(&retry.TaskRetry{
		MaxRetries: s.config.Persistence.MaxRetries,
		Backoff:    true,
	}, func() error {
		s.lock.RLock()
		err := s.kv.RefreshLease(s.frameworkLease)
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to refresh framework ID lease: %s", err.Error())
		}

		s.lock.RUnlock()

		return err
	})
}

func (s *SprintEventController) getAllTasks() (map[string]string, error) {
	var tasks map[string]string
	err := s.taskmanager.RunPolicy(&retry.TaskRetry{
		MaxRetries: s.config.Persistence.MaxRetries,
		Backoff:    true,
	}, func() error {
		t, err := s.kv.ReadAll("/tasks")
		if err != nil {
			s.logger.Emit(logging.ERROR, "Failed to get all task data: %s", err.Error())
			return err
		}

		tasks = t
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

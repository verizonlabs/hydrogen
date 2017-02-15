package scheduler

import ()
import (
	exec "mesos-framework-sdk/include/executor"
	mesos "mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-sdk/backoff"
	"mesos-sdk/extras"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	"sprint/scheduler/taskmanager"
)

// Returns a new ExecutorInfo with a new UUID.
// Reuses existing data from the scheduler's original ExecutorInfo.
func (s *SprintScheduler) NewExecutor() *mesos.ExecutorInfo {
	uuid, _ := extras.UuidToString(extras.Uuid())

	return &mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{Value: uuid},
		Name:       s.executor.Name,
		Command:    s.executor.Command,
		Resources:  s.executor.Resources,
		Container:  s.executor.Container,
	}
}

// Returns the scheduler's configuration.
func (s *SprintScheduler) Config() configuration {
	return s.config
}

// Returns the scheduler's configuration.
func (s *SprintScheduler) TaskManager() *taskmanager.Manager {
	return s.taskmgr
}

// Returns the internal state of the scheduler
func (s *SprintScheduler) State() *state {
	return &s.state
}

// Returns the caller that we use for communication.
func (s *SprintScheduler) Caller() *calls.Caller {
	return &s.http
}

// Returns the FrameworkInfo that is sent to Mesos.
func (s *SprintScheduler) FrameworkInfo() *mesos.FrameworkInfo {
	return s.framework
}

// Returns the ExecutorInfo that's associated with the scheduler.
func (s *SprintScheduler) ExecutorInfo() *mesos.ExecutorInfo {
	return s.executor
}

// Runs our scheduler with some applied configuration.
func (s *SprintScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return c.Run(*config)
}

// This call suppresses our offers received from Mesos.
func (s *SprintScheduler) SuppressOffers() error {
	return calls.CallNoData(s.http, calls.Suppress())
}

// This call will perform reconciliation for our tasks.
func (s *SprintScheduler) Reconcile() (mesos.Response, error) {
	tasks, err := taskmanager.TaskIdAgentIdMap(s.taskmgr.Tasks())
	if err != nil {
		mlog.Error(err.Error())
	}
	return calls.Caller(s.http).Call(
		calls.Reconcile(
			calls.ReconcileTasks(tasks),
		),
	)
}

// This call revives our offers received from Mesos.
func (s *SprintScheduler) ReviveOffers() error {
	select {
	// Rate limit offer revivals.
	case <-s.state.reviveTokens:
		return calls.CallNoData(s.http, calls.Revive())
	default:
		return nil
	}
}

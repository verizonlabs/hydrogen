package plan

import (
	scheduler2 "github.com/verizonlabs/hydrogen/scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	resource "github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"hydrogen/task/persistence"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/scheduler"
	"time"
)

type ReconcilePlan struct {
	Plan
	tasks           []*manager.Task
	taskManager     manager.TaskManager
	resourceManager resource.ResourceManager
	scheduler       scheduler.Scheduler
	storage         persistence.Storage
	config          scheduler2.Configuration
	state           PlanState
	logger          logging.Logger
}

func NewReconcilePlan(tasks []*manager.Task,
	taskManager manager.TaskManager,
	resourceManager resource.ResourceManager,
	scheduler scheduler.Scheduler,
	storage persistence.Storage,
	config scheduler2.Configuration,
	logger logging.Logger) Plan {
	return ReconcilePlan{
		taskManager:     taskManager,
		resourceManager: resourceManager,
		scheduler:       scheduler,
		storage:         storage,
		config:          config,
		state:           QUEUED,
		logger:          logger,
	}
}

func (r *ReconcilePlan) Execute() error {
	return nil
} // Execute the current plan.

func (r *ReconcilePlan) Update() error {
	return nil
} // Update the current plan status.

func (r *ReconcilePlan) Status() PlanState {
	return r.state
} // Current status of the plan.

func (r *ReconcilePlan) Type() PlanType {
	return Reconcile
} // Tells you the plan type.

// Keep our state in check by periodically reconciling.
func (r *ReconcilePlan) periodicReconcile() {
	ticker := time.NewTicker(r.config.Scheduler.ReconcileInterval)
	for {
		select {
		case <-ticker.C:
			recon, err := r.taskManager.AllByState(manager.RUNNING)
			if err != nil {
				continue
			}
			toReconcile := []*mesos_v1.TaskInfo{}
			for _, t := range recon {
				toReconcile = append(toReconcile, t.Info)
			}
			_, err = r.scheduler.Reconcile(toReconcile)
			if err != nil {
				r.logger.Emit(logging.ERROR, "Failed to reconcile all running tasks: %s", err.Error())
			}
		}
	}
}

package events

import "mesos-framework-sdk/logging"

func (d *SprintExecutorController) Shutdown() {
	d.logger.Emit(logging.INFO, "Executor is shutting down...")
}

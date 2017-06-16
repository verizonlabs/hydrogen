package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/utils"
	"testing"
)

func TestSprintEventController_Update(t *testing.T) {
	ctrl := workingEventController()
	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
				State:  state,
			}})
	}
}

func TestSprintEventController_UpdateWithNilTaskId(t *testing.T) {
	ctrl := workingEventController()
	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: nil},
				State:  state,
			}})
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: nil,
				State:  state,
			}})
	}
	ctrl.Update(nil)
}

func TestSprintEventController_UpdateWithInvalidState(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Update(&mesos_v1_scheduler.Event_Update{
		Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: nil},
			State:  nil,
		}})
	ctrl.Update(&mesos_v1_scheduler.Event_Update{
		Status: &mesos_v1.TaskStatus{
			TaskId: &mesos_v1.TaskID{Value: nil},
			State:  mesos_v1.TaskState(100).Enum(), // Doesn't exist.
		}})
}

func TestSprintEventController_UpdateWith(t *testing.T) {
	ctrl := brokenSchedulerEventController()
	for _, state := range states {
		ctrl.Update(&mesos_v1_scheduler.Event_Update{
			Status: &mesos_v1.TaskStatus{
				TaskId: &mesos_v1.TaskID{Value: utils.ProtoString("id")},
				State:  state,
			}})
	}
}

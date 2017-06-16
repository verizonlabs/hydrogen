package events

import "testing"

// Can we create a leader?
func TestSprintEventController_CreateLeader(t *testing.T) {
	ctrl := workingEventController()
	if err := ctrl.CreateLeader(); err != nil {
		t.Logf("Failed to create a leader %v", err.Error())
		t.Fail()
	}
}

// Can we create and get the leader?
func TestSprintEventController_GetLeader(t *testing.T) {
	ctrl := workingEventController()
	if err := ctrl.CreateLeader(); err != nil {
		t.Logf("Failed to create a leader %v\n", err.Error())
		t.Fail()
	}
	// Leader string is simply an empty string during testing...
	_, err := ctrl.GetLeader()
	if err != nil {
		t.Logf("Failed to grab the leader %v\n", err.Error())
		t.Fail()
	}
}

package events

import "testing"

func TestSprintEventController_Communicate(t *testing.T) {
	ctrl := workingEventControllerFactory()
	go ctrl.Communicate() // this never gets covered since this test closes too fast
	ctrl.Election()
}

func TestSprintEventController_CreateAndGetLeader(t *testing.T) {
	ctrl := workingEventControllerFactory()
	ctrl.CreateLeader()
	ctrl.GetLeader()
}

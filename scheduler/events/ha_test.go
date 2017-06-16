package events

import "testing"

// Possible to test communication with a variadic input for a channel.
// If a channel is passed in, it communicates on that channel.
// Can block on the "read" of that channel.
func TestSprintEventController_Communicate(t *testing.T) {
	ctrl := workingEventController()
	go ctrl.Communicate() // this never gets covered since this test closes too fast
	ctrl.Election()
}

func TestSprintEventController_Election(t *testing.T) {
	ctrl := workingEventController()
	ctrl.Election()
}

package ha

import (
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence"
	"mesos-framework-sdk/scheduler/events"
	"net"
	"sprint/scheduler"
	sprintEvents "sprint/scheduler/events"
)

// Performs leader election or becomes a standby if it's determined that there's already a leader.
func LeaderElection(c *scheduler.Configuration, s events.SchedulerEvent, kv persistence.Storage, l logging.Logger) {
	e := s.(*sprintEvents.SprintEventController)

	for {
		// This will only set us as the leader if there isn't an already existing leader.
		e.CreateLeader()

		leader := e.GetLeader()
		if leader != c.Leader.IP {
			l.Emit(logging.INFO, "Connecting to leader to determine when we need to wake up and perform leader election")

			// Block here until we lose connection to the leader.
			// Once the connection has been lost elect a new leader.
			err := LeaderClient(c, leader)

			// Only delete the key if we've lost the connection, not timed out.
			// This conditional requires Go 1.6+
			if err, ok := err.(net.Error); ok && err.Timeout() {
				l.Emit(logging.ERROR, "Timed out connecting to leader")
			} else {
				l.Emit(logging.ERROR, "Lost connection to leader")
				kv.Delete("/leader")
			}
		} else {

			// We are the leader, exit the loop and start the scheduler/API.
			break
		}
	}
}

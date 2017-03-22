package ha

import (
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence"
	"net"
	"sprint/scheduler"
	"sprint/scheduler/events"
	"time"
)

// Performs leader election or becomes a standby if it's determined that there's already a leader.
func LeaderElection(c *scheduler.SchedulerConfiguration, e *events.SprintEventController, kv persistence.KVStorage, l logging.Logger) {
	for {

		// This will only set us as the leader if there isn't an already existing leader.
		e.CreateLeader()

		leader, err := e.GetLeader()
		if err != nil {
			l.Emit(logging.ERROR, "Couldn't get leader: %s", err.Error())
			time.Sleep(c.LeaderRetryInterval)
			continue
		}

		if err != nil {
			l.Emit(logging.ERROR, "Couldn't determine IPs for interface: %s", err.Error())
			time.Sleep(c.LeaderRetryInterval)
			continue
		}

		if leader != c.LeaderIP {
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

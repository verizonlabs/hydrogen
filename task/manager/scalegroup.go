package manager

import (
	"mesos-framework-sdk/include/mesos_v1"
)

type (
	ScaleGrouping interface {
		CreateGroup(string) error               // Creates a new task group
		DeleteGroup(string) error               // Deletes an entire task group
		SetSize(string, int) error              // Can handle positive/negative values.
		Link(string, *mesos_v1.AgentID) error   // Links an agent from a task group.
		Unlink(string, *mesos_v1.AgentID) error // Unlinks an agent from a task group
		ReadGroup(string) []*mesos_v1.AgentID   // Gathers all agents tied to a task group
		IsInGroup(*mesos_v1.TaskInfo) bool      // Checks for existence of a taskgroup
	}
)

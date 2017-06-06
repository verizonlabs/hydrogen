package manager

import (
	"mesos-framework-sdk/include/mesos_v1"
	"sprint/task/persistence"
)

type (
	ScaleGrouping interface {
		CreateGroup(string) error
		AddToGroup(string, *mesos_v1.AgentID) error
		ReadGroup(string) []*mesos_v1.AgentID
		DelFromGroup(string, *mesos_v1.AgentID) error
		DeleteGroup(string) error
		IsInGroup(*mesos_v1.TaskInfo) bool
	}
	ScaleGroup struct {
		name    string
		backend persistence.Storage
	}
)

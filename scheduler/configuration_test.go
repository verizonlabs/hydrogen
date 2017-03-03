package scheduler

import (
	"os/user"
	"testing"
	"time"
)

var sprintConfig = new(SchedulerConfiguration).Initialize()

// Make sure we return the right name.
func TestSchedulerConfiguration_Name(t *testing.T) {
	t.Parallel()

	if sprintConfig.Name() != "Sprint" {
		t.Fatal("Configuration has wrong name")
	}
}

// Ensures that we can detect the current user and pass it into the framework info.
func TestSchedulerConfiguration_User(t *testing.T) {
	t.Parallel()

	u, err := user.Current()
	if err != nil {
		t.Fatal("Unable to detect current user: " + err.Error())
	}

	if sprintConfig.User() != u.Username {
		t.Fatal("User is not set correctly")
	}
}

// Checks to see if our default value for checkpointing is right.
func TestSchedulerConfiguration_Checkpointing(t *testing.T) {
	t.Parallel()

	if !*sprintConfig.Checkpointing() {
		t.Fatal("Checkpointing is not set to the right value")
	}
}

// Make sure we have the right principal value.
func TestSchedulerConfiguration_Principal(t *testing.T) {
	t.Parallel()

	if sprintConfig.Principal() != "Sprint" {
		t.Fatal("Principal is not set to the right value")
	}
}

// Checks to see whether the command is set properly.
func TestSchedulerConfiguration_Command(t *testing.T) {
	t.Parallel()

	if *sprintConfig.Command() != "" {
		t.Fatal("Command is not set to the right value")
	}
}

// Make sure the URIs are set correctly.
func TestSchedulerConfiguration_Uris(t *testing.T) {
	t.Parallel()

	if len(sprintConfig.Uris()) != 0 {
		t.Fatal("The number of URIs should be 0")
	}
}

// Check our default timeout value.
func TestSchedulerConfiguration_Timeout(t *testing.T) {
	t.Parallel()

	if sprintConfig.PersistenceTimeout() != time.Second {
		t.Fatal("Timeout is not set to the right value")
	}
}

// Make sure we have the right default endpoint value.
func TestSchedulerConfiguration_Endpoint(t *testing.T) {
	t.Parallel()

	if sprintConfig.Endpoint() != "http://127.0.0.1:5050/api/v1/scheduler" {
		t.Fatal("Endpoint is not set to the right value")
	}
}

// Ensure we have the right revive burst amount.
func TestSchedulerConfiguration_ReviveBurst(t *testing.T) {
	t.Parallel()

	if sprintConfig.ReviveBurst() != 3 {
		t.Fatal("Revive burst is not set to the right value")
	}
}

// Ensure we have the right revive wait period.
func TestSchedulerConfiguration_ReviveWait(t *testing.T) {
	t.Parallel()

	if sprintConfig.ReviveWait() != 1*time.Second {
		t.Fatal("Revive wait period is not set to the right value")
	}
}

// Ensure we have the right maximum refusal time.
func TestSchedulerConfiguration_MaxRefuse(t *testing.T) {
	t.Parallel()

	if sprintConfig.MaxRefuse() != 5*time.Second {
		t.Fatal("Max refusal time is not set to the right value")
	}
}

// Make sure we get our executor name correctly.
func TestSchedulerConfiguration_ExecutorName(t *testing.T) {
	t.Parallel()

	if *sprintConfig.ExecutorName() != "Sprinter" {
		t.Fatal("Executor name is wrong")
	}
}

// Make sure we get our executor command correctly.
func TestSchedulerConfiguration_ExecutorCmd(t *testing.T) {
	t.Parallel()

	if *sprintConfig.ExecutorCmd() != "./executor" {
		t.Fatal("Executor command is wrong")
	}
}

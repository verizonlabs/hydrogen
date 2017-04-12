package builder

import (
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/logging/test"
	"mesos-framework-sdk/task"
	"testing"
)

func TestApplication(t *testing.T) {
	a := make([]map[string]string, 0)
	b := make([]task.Filter, 0)
	// Test
	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
		},
		Command: &task.CommandJSON{
			Cmd: proto.String("/bin/sleep 1"),
		},
		Container:   &task.ContainerJSON{},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationNoName(t *testing.T) {
	a := make([]map[string]string, 0)
	b := make([]task.Filter, 0)
	// Test
	test := &task.ApplicationJSON{
		Name:        "",
		Resources:   &task.ResourceJSON{},
		Command:     &task.CommandJSON{},
		Container:   &task.ContainerJSON{},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationNoResources(t *testing.T) {
	a := make([]map[string]string, 0)
	b := make([]task.Filter, 0)
	// Test
	test := &task.ApplicationJSON{
		Name:        "Test task",
		Resources:   nil,
		Command:     &task.CommandJSON{},
		Container:   &task.ContainerJSON{},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationCommandFail(t *testing.T) {
	a := make([]map[string]string, 0)
	b := make([]task.Filter, 0)
	// Test
	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
		},
		Command: &task.CommandJSON{
			Cmd: nil,
		},
		Container:   &task.ContainerJSON{},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationContainerFail(t *testing.T) {
	a := make([]map[string]string, 0)
	b := make([]task.Filter, 0)
	l := make([]task.VolumesJSON, 1)
	l = append(l, task.VolumesJSON{HostPath: proto.String("/home/someone")})
	// Test
	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
		},
		Command: &task.CommandJSON{
			Cmd: proto.String("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     nil,
			ContainerType: proto.String("docker"),
			Volumes:       l,
		},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)
	if err == nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestApplicationWithDockerContainer(t *testing.T) {
	a := make([]map[string]string, 0)
	b := make([]task.Filter, 0)

	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
		},
		Command: &task.CommandJSON{
			Cmd: proto.String("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     proto.String("debian:latest"),
			ContainerType: proto.String("docker"),
		},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationFailResources(t *testing.T) {
	a := make([]map[string]string, 0)
	b := make([]task.Filter, 0)

	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0,
			Mem: 0,
		},
		Command: &task.CommandJSON{
			Cmd: proto.String("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     proto.String("debian:latest"),
			ContainerType: proto.String("docker"),
		},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationFailLabels(t *testing.T) {
	a := make([]map[string]string, 1)
	b := make([]task.Filter, 0)
	a = append(a, map[string]string{"": ""})

	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 1.0,
			Mem: 128.0,
		},
		Command: &task.CommandJSON{
			Cmd: proto.String("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     proto.String("debian:latest"),
			ContainerType: proto.String("docker"),
		},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	lgr := new(mockLogging.MockLogger)
	_, err := Application(test, lgr)

	if err == nil {
		t.FailNow()
	}
}

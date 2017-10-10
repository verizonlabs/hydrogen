// Copyright 2017 Verizon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builder

import (
	"github.com/verizonlabs/mesos-framework-sdk/task"
	"github.com/verizonlabs/mesos-framework-sdk/utils"
	"testing"
)

func TestApplication(t *testing.T) {
	a := make(map[string]string, 0)
	b := make([]task.Filter, 0)
	// Test
	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
			Disk: task.Disk{
				Size: 1024.0,
			},
		},
		Command: &task.CommandJSON{
			Cmd: utils.ProtoString("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{},
		HealthCheck: &task.HealthCheckJSON{
			Type: utils.ProtoString("http"),
			Http: &task.HTTPHealthCheck{
				Scheme: utils.ProtoString("http"),
				Port:   utils.ProtoInt32(9090),
				Path:   utils.ProtoString("/somepath"),
				Statuses: []uint32{
					200,
					202,
				},
			},
			DelaySeconds: utils.ProtoFloat64(1.0),
		},
		Labels:  a,
		Filters: b,
	}
	_, err := Application(test)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationNoName(t *testing.T) {
	a := make(map[string]string, 0)
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
	_, err := Application(test)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationNoResources(t *testing.T) {
	a := make(map[string]string, 0)
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
	_, err := Application(test)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationCommandFail(t *testing.T) {
	a := make(map[string]string, 0)
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
	_, err := Application(test)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationContainerFail(t *testing.T) {
	a := make(map[string]string, 0)
	b := make([]task.Filter, 0)
	l := make([]task.VolumesJSON, 1)
	l = append(l, task.VolumesJSON{HostPath: utils.ProtoString("/home/someone")})
	// Test
	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
		},
		Command: &task.CommandJSON{
			Cmd: utils.ProtoString("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     nil,
			ContainerType: utils.ProtoString("docker"),
			Volumes:       l,
		},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	_, err := Application(test)
	if err == nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestApplicationWithDockerContainer(t *testing.T) {
	a := make(map[string]string, 0)
	b := make([]task.Filter, 0)

	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
			Disk: task.Disk{
				Size: 1024.0,
			},
		},
		Command: &task.CommandJSON{
			Cmd: utils.ProtoString("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     utils.ProtoString("debian:latest"),
			ContainerType: utils.ProtoString("docker"),
		},
		HealthCheck: &task.HealthCheckJSON{
			Type: utils.ProtoString("http"),
			Http: &task.HTTPHealthCheck{
				Scheme: utils.ProtoString("http"),
				Port:   utils.ProtoInt32(9090),
				Path:   utils.ProtoString("/somepath"),
				Statuses: []uint32{
					200,
					202,
				},
			},
			DelaySeconds: utils.ProtoFloat64(1.0),
		},
		Labels:  a,
		Filters: b,
	}
	_, err := Application(test)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationFailResources(t *testing.T) {
	a := make(map[string]string, 0)
	b := make([]task.Filter, 0)

	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0,
			Mem: 0,
		},
		Command: &task.CommandJSON{
			Cmd: utils.ProtoString("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     utils.ProtoString("debian:latest"),
			ContainerType: utils.ProtoString("docker"),
		},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	_, err := Application(test)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestApplicationFailLabels(t *testing.T) {
	a := map[string]string{"": ""}
	b := make([]task.Filter, 0)

	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 1.0,
			Mem: 128.0,
		},
		Command: &task.CommandJSON{
			Cmd: utils.ProtoString("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{
			ImageName:     utils.ProtoString("debian:latest"),
			ContainerType: utils.ProtoString("docker"),
		},
		HealthCheck: &task.HealthCheckJSON{},
		Labels:      a,
		Filters:     b,
	}
	_, err := Application(test)

	if err == nil {
		t.FailNow()
	}
}

func TestApplicationHealthChecks(t *testing.T) {
	a := make(map[string]string, 0)
	b := make([]task.Filter, 0)

	// Test
	test := &task.ApplicationJSON{
		Name: "Test Task",
		Resources: &task.ResourceJSON{
			Cpu: 0.5,
			Mem: 128.0,
			Disk: task.Disk{
				Size: 1024.0,
			},
		},
		Command: &task.CommandJSON{
			Cmd: utils.ProtoString("/bin/sleep 1"),
		},
		Container: &task.ContainerJSON{},
		HealthCheck: &task.HealthCheckJSON{
			Type: utils.ProtoString("http"),
			Http: &task.HTTPHealthCheck{
				Scheme: utils.ProtoString("http"),
				Port:   utils.ProtoInt32(9090),
				Path:   utils.ProtoString("/somepath"),
				Statuses: []uint32{
					200,
					202,
				},
			},
			DelaySeconds: utils.ProtoFloat64(1.0),
		},
		Labels:  a,
		Filters: b,
	}
	_, err := Application(test)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

package manager

import (
	"testing"
	k "mesos-framework-sdk/resources/manager/test"
	s "mesos-framework-sdk/scheduler/test"
	"sprint/task/manager/test"
	"mesos-framework-sdk/include/mesos_v1"
)

// Generate valid and invalid JSON

func TestNewApiParser(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	if api.resourceManager == nil || api.scheduler == nil || api.taskManager == nil {
		t.Logf("Expected instances to be set %v\n", api)
		t.Fail()
	}
}

func TestParser_DeployNoHealthCheck(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	validJSON := `{"name": "test",
	"instances": 1,
	"resources": {"cpu": 0.5, "mem": 128.0, "disk": {"size": 1024.0}},
	"command": {"cmd": "echo hello"}}`

	task, err := api.Deploy([]byte(validJSON))
	if err != nil {
		t.Logf("Failure to parse JSON %v\n", err)
		t.Fail()
	}
	if task.HealthCheck != nil {
		t.Log("Healthcheck field was supposed to be set as nil, was non-nil instead.")
		t.Fail()
	}
}

func TestParser_DeployWithTCPHealthCheck(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	validJSON := `{"name": "test",
	"instances": 1,
	"resources": {"cpu": 0.5, "mem": 128.0, "disk": {"size": 1024.0}},
	"command": {"cmd": "echo hello"},
	"healthcheck": {
	    "type": "tcp",
	    "tcp": {
	      "port": 9000
	    }
	  }
	}`

	task, err := api.Deploy([]byte(validJSON))
	if err != nil {
		t.Logf("Failure to parse JSON %v\n", err)
		t.Fail()
	}
	if task == nil {
		t.Log("Task is nil and should be set to some value.")
		t.Fail()
	} else if task.HealthCheck == nil {
		t.Log("Healthcheck field was supposed to be set, was nil instead.")
		t.Fail()
	} else if task.HealthCheck.Type == nil{
		t.Log("Healthcheck type was set to nil, and not tcp")
		t.Fail()
	}
	if task.HealthCheck.Type.String() != "TCP" {
		t.Logf("Healthcheck type was set to %v instead of tcp", task.HealthCheck.Type.String())
		t.Failed()
	}
}

func TestParser_DeployWithNoName(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	invalidJSON := `{"instances": 1,
	"resources": {"cpu": 0.5, "mem": 128.0, "disk": {"size": 1024.0}},
	"command": {"cmd": "echo hello"}`
	task, err := api.Deploy([]byte(invalidJSON))
	if err == nil {
		t.Logf("Failure to determine unnamed task %v, must have a name\n", task)
		t.Fail()
	}
}

func TestParser_DeployWithNoResources(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	invalidJSON := `{"name": "no-resources",
	"instances": 1,
	"command": {"cmd": "echo hello"}`
	task, err := api.Deploy([]byte(invalidJSON))
	if err == nil {
		t.Logf("Failure to determine task with no resources %v, must have set resources.\n", task)
		t.Fail()
	}
}

func TestParser_DeployWithCNINetwork(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	validJSON := `{"name": "tester",
	"instances": 1,
	"resources": {"cpu": 0.5, "mem": 128.0, "disk": {"size": 1024.0}},
	"container": {
		"image": "debian:latest",
		"network": [{"name": "cni"}]
	},
	"command": {"cmd": ""}
	}`
	task, err := api.Deploy([]byte(validJSON))
	if err != nil {
		t.Logf("Error..%v\n", err.Error())
		t.Logf("Failure to determine task with CNI %v, CNI should be set\n", task)
		t.Fail()
	}
}

func TestParser_DeployWithIPNetwork(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	validJSON := `{"name": "tester",
	"instances": 1,
	"resources": {"cpu": 0.5, "mem": 128.0, "disk": {"size": 1024.0}},
	"container": {
		"image": "debian:latest",
		"network": [{
			"ipaddress": [{
					"ip": "10.2.1.25",
					"protocol": "ipv4"
			}],
			"group": ["test", "stuff"],
			"labels": [{"some": "label"}]
		}]
	},
	"command": {"cmd": ""}
	}`
	task, err := api.Deploy([]byte(validJSON))
	if err != nil {
		t.Logf("Error parsing JSON for IP address setting %v\n", err.Error())
		t.Fail()
	}
	if task.GetContainer() == nil {
		t.Logf("Container is nil, should be set %v\n", task.GetContainer())
		t.Fail()
	} else if len(task.GetContainer().GetNetworkInfos()) == 0 {
		t.Logf("Container networking list is empty %v\n", task.GetContainer().GetNetworkInfos())
		t.Fail()
	}
	net := task.GetContainer().GetNetworkInfos()[0]
	if len(net.IpAddresses) > 1 {
		t.Log("Container network has > 1 ip address, only expecting 1.")
		t.Fail()
	} else if net.IpAddresses[0].GetIpAddress() != "10.2.1.25" {
		t.Logf("Container networking has the wrong IP %v\n", task.GetContainer().GetNetworkInfos()[0])
		t.Fail()
	} else if len(net.Groups) > 2 || len(net.Groups) < 2{
		t.Logf("Expecting only 2 groups %v\n", net.Groups)
		t.Fail()
	} else if len(net.Labels.Labels) != 1 {
		t.Logf("Expecting only a single key value pair for labels %v\n", net.Labels)
		t.Fail()
	} else if net.Name != nil {
		t.Logf("Expecting name for the network to be empty, was not nil %v", net.Name)
		t.Fail()
	}
}

func TestParser_Kill(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	validJSON := `{"name": "test"}`
	status, err := api.Kill([]byte(validJSON))
	if err != nil {
		t.Logf("Failed %v\n", err)
		t.Fail()
	}
	if status == "" {
		t.Logf("Replied killed with an empty status! %v\n", status)
		t.Fail()
	}
	if status != "test" {
		t.Logf("Name of app killed was not 'test', instead got %v\n", status)
		t.Fail()
	}
}

func TestParser_KillFail(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	validJSON := `{"junk":"value"}`
	status, err := api.Kill([]byte(validJSON))
	if err == nil {
		t.Logf("Application should of failed %v\n", err)
		t.Fail()
	}
	if status != "" {
		t.Logf("Kill was supposed to fail given a junk value %v\n", status)
	}
}

func TestParser_AllTasks(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	tasks, err := api.AllTasks()
	if err != nil {
		t.Logf("Failed %v\n", err)
		t.Fail()
	}
	if len(tasks) != 1 {
		t.Logf("Tasks length should be 1, found %v", len(tasks))
		t.Fail()
	}
}

func TestParser_Update(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	validJSON := `{"name": "test",
	"instances": 1,
	"resources": {"cpu": 0.5, "mem": 128.0, "disk": {"size": 1024.0}},
	"command": {"cmd": "echo hello"}}`
	task, err := api.Update([]byte(validJSON))
	if err != nil {
		t.Logf("Failed %v\n", err)
		t.Fail()
	}
	if task.GetName() != "test" {
		t.Logf("Task updated came back with different name %v, should be the same", task.GetName())
	}
}

func TestParser_Status(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	state, err := api.Status("test")
	if err != nil {
		t.Logf("Failed on status update %v\n", state)
		t.Fail()
	}
	if state.String() != mesos_v1.TaskState_TASK_STARTING.String() {
		t.Logf("Expected task running, got %v", state.String())
		t.Fail()
	}
}

func TestParser_DeployMultiInstance(t *testing.T) {
	api := NewApiParser(k.MockResourceManager{}, test.MockTaskManager{}, s.MockScheduler{})
	multiInstance := `{"name": "test",
	"instances": 5,
	"resources": {"cpu": 0.5, "mem": 128.0, "disk": {"size": 1024.0}},
	"command": {"cmd": "echo hello"}}`
	task, err := api.Deploy([]byte(multiInstance))
	if err != nil {
		t.Logf("Deploying multiple instances failed %v", task)
		t.Fail()
	}

}
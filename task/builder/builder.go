package builder

import (
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	resourcebuilder "mesos-framework-sdk/resources"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/utils"
)

func Application(t *task.ApplicationJSON, lgr *logging.DefaultLogger) *mesos_v1.TaskInfo {
	// Allocate space for our resources.
	var resources []*mesos_v1.Resource
	var cpu = resourcebuilder.CreateCpu(t.Resources.Cpu, t.Resources.Role)
	var mem = resourcebuilder.CreateMem(t.Resources.Mem, t.Resources.Role)

	networks, err := task.ParseNetworkJSON(t.Container.Network)
	if err != nil {
		// This isn't a fatal error so we can log this as debug and move along.
		lgr.Emit(logging.INFO, "No explicit network info passed in.")
	}

	resources = append(resources, cpu, mem)

	uuid, err := utils.UuidToString(utils.Uuid())
	if err != nil {
		lgr.Emit(logging.ERROR, err.Error())
	}

	container := resourcebuilder.CreateContainerInfoForMesos(
		resourcebuilder.CreateImage(
			*t.Container.ImageName, "", mesos_v1.Image_DOCKER.Enum(),
		),
	)

	return resourcebuilder.CreateTaskInfo(
		proto.String(t.Name),
		&mesos_v1.TaskID{Value: proto.String(uuid)},
		resourcebuilder.CreateSimpleCommandInfo(t.Command.Cmd, nil),
		resources,
		resourcebuilder.CreateMesosContainerInfo(container, networks),
	)
}

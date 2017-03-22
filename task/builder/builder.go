package builder

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	resourcebuilder "mesos-framework-sdk/resources"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/network"
	"mesos-framework-sdk/task/volume"
	"mesos-framework-sdk/utils"
)

func Application(t *task.ApplicationJSON, lgr logging.Logger) (*mesos_v1.TaskInfo, error) {
	// Allocate space for our resources.
	resources := make([]*mesos_v1.Resource, 0)
	var cpu = resourcebuilder.CreateCpu(t.Resources.Cpu, t.Resources.Role)
	var mem = resourcebuilder.CreateMem(t.Resources.Mem, t.Resources.Role)

	networks, err := network.ParseNetworkJSON(t.Container.Network)
	if err != nil {
		// This isn't a fatal error so we can log this as debug and move along.
		lgr.Emit(logging.INFO, "No explicit network info passed in, using default host networking.")
	}
	vol := make([]*mesos_v1.Volume, 0)
	if len(t.Container.Volumes) > 0 {
		vol, err = volume.ParseVolumeJSON(t.Container.Volumes)
		if err != nil {
			lgr.Emit(logging.ERROR, err.Error())
			return nil, errors.New("Error parsing volume JSON: " + err.Error())
		}
	}

	resources = append(resources, cpu, mem)

	uuid, err := utils.UuidToString(utils.Uuid())
	if err != nil {
		lgr.Emit(logging.ERROR, err.Error())
	}

	var container *mesos_v1.ContainerInfo_MesosInfo
	if t.Container.ImageName != nil {
		container = resourcebuilder.CreateContainerInfoForMesos(
			resourcebuilder.CreateImage(
				*t.Container.ImageName, "", mesos_v1.Image_DOCKER.Enum(),
			),
		)
	} else {
		return nil, errors.New("Container image name was not passed in. Please pass in a container name.")
	}

	u := make([]*mesos_v1.CommandInfo_URI, 0)

	if t.Command != nil && len(t.Command.Uris) > 0 {
		for _, uri := range t.Command.Uris {
			t := &mesos_v1.CommandInfo_URI{
				Value:      uri.Uri,
				Executable: uri.Execute,
				Extract:    uri.Extract,
			}
			u = append(u, t)
		}
	}

	labels := make([]*mesos_v1.Label, 0)
	if t.Labels != nil {
		for _, labelList := range t.Labels {
			for k, v := range labelList {
				label := &mesos_v1.Label{
					Key:   proto.String(k),
					Value: proto.String(v),
				}
				labels = append(labels, label)
			}
		}
	}
	labelProto := &mesos_v1.Labels{Labels: labels}

	return resourcebuilder.CreateTaskInfo(
		proto.String(t.Name),
		&mesos_v1.TaskID{Value: proto.String(uuid)},
		resourcebuilder.CreateSimpleCommandInfo(t.Command.Cmd, u),
		resources,
		resourcebuilder.CreateMesosContainerInfo(container, networks, vol, nil),
		labelProto,
	), nil
}

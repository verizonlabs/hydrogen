package executor

import "mesos-sdk"

/*
	Returns a commandInfo protobuf with the specified command
*/
func CommandInfo(command string) mesos.CommandInfo {
	return mesos.CommandInfo{
		Value: ProtoString(command),
		// URI is needed to pull the executor down!
		URIs: []mesos.CommandInfo_URI{
			{
				Value:      executorFetcherURI,
				Executable: ProtoBool(isExecutable),
			},
		},
	}
}

func ProtoString(s string) *string { return &s }

func ProtoBool(b bool) *bool { return &b }

/*
	Convenience function to get a mesos container running a docker image.
	Requires slave to run with mesos containerization flag set.
*/
func Container(imageName string) *mesos.ContainerInfo {
	return &mesos.ContainerInfo{
		Type: mesos.ContainerInfo_MESOS.Enum(),
		Mesos: &mesos.ContainerInfo_MesosInfo{
			Image: &mesos.Image{
				Docker: &mesos.Image_Docker{
					Name: imageName,
				},
				Type: mesos.Image_DOCKER.Enum(),
			},
		},
	}
}

/*
	Convenience function to return mesos CPU resources.
	CPU resources can be fractional.
*/
func CpuResources(amount float64) mesos.Resource {
	return mesos.Resource{
		Name:   "cpus",
		Type:   mesos.SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: 0.5},
	}
}

/*
	Convenience function to return mesos memory resources.
	Memory allocated is in mb.
*/
func MemResources(amount float64) mesos.Resource {
	return mesos.Resource{
		Name:   "mem",
		Type:   mesos.SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: amount},
	}
}

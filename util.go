package sprint

import "mesos-sdk"

// Returns a string reference.
func ProtoString(s string) *string { return &s }

// Returns a boolean reference.
func ProtoBool(b bool) *bool { return &b }

// Convenience function to return mesos CPU resources.
// CPU resources can be fractional.
func CpuResources(amount float64) mesos.Resource {
	return mesos.Resource{
		Name:   "cpus",
		Type:   mesos.SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: 0.5},
	}
}

// Convenience function to return mesos memory resources.
// Memory allocated is in mb.
func MemResources(amount float64) mesos.Resource {
	return mesos.Resource{
		Name:   "mem",
		Type:   mesos.SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: amount},
	}
}

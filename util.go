package sprint

import "mesos-sdk"

// Returns a string reference.
func ProtoString(s string) *string { return &s }

// Returns a boolean reference.
func ProtoBool(b bool) *bool { return &b }

// Convenience function to set resources.
func Resource(name string, amount float64) mesos.Resource {
	return mesos.Resource{
		Name:   name,
		Type:   mesos.SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: amount},
	}
}

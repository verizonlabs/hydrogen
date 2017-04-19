package main

import (
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/executor"
	"mesos-framework-sdk/include/executor"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/utils"
	"net"
	"os"
	"sprint/executor/events"
)

var (
	IPv4Bits = 32
	Subnet   = 24
)

// Gathers the internal network as defined.
// This will not work if there are multiple
// 10.0.0.0/24's on a host.
// NOTE (tim): Talk to mike c about how the new
// /25 networks will work, as well as overlay
// networks.
func getInternalNetworkInterface() (net.IP, error) {
	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, interFace := range interfaces {
		ip, net, err := net.ParseCIDR(interFace.String())
		if err != nil {
			return nil, err
		}
		ones, bits := net.Mask.Size()
		// If it's v4
		if bits <= IPv4Bits {
			// Is this a /24 network?
			if ones == Subnet {
				// IP is padded to the left for ipv6.
				// First bit for ipv4 starts at 12th index.
				// NOTE (tim): This magic 10 number will
				// have to change going forward to support
				// new networking architecture.
				if ip[12] == byte(10) {
					return ip, nil
				}
			}
		}
	}
	return nil, nil
}

// Main function will wire up all other dependencies for the executor and setup top-level configuration.
func main() {
	logger := logging.NewDefaultLogger()
	// Set implicitly by the Mesos agent.
	fwId := &mesos_v1.FrameworkID{Value: utils.ProtoString(os.Getenv("MESOS_FRAMEWORK_ID"))}
	execId := &mesos_v1.ExecutorID{Value: utils.ProtoString(os.Getenv("MESOS_EXECUTOR_ID"))}

	// If an ENV VAR was set for an endpoint, we use that.
	endpoint := os.Getenv("EXECUTOR_ENDPOINT")

	// Otherwise we find the main interface as specified.
	if endpoint == "" {
		ip, err := getInternalNetworkInterface()
		if err != nil {
			logger.Emit(logging.ERROR, "Unable to find an internal IP address, "+
				"subnet size %v and address family of %v bits", Subnet, IPv4Bits)
			os.Exit(1)
		}
		endpoint = ip.String()
	}

	port := os.Getenv("EXECUTOR_ENDPOINT_PORT")
	if port == "" {
		// Mesos default executor port.
		port = "5051"
	}

	logger.Emit(logging.INFO, "Endpoint set to "+"http://"+endpoint+":"+port+"/api/v1/executor")
	c := client.NewClient("http://"+endpoint+":"+port+"/api/v1/executor", logger)
	ex := executor.NewDefaultExecutor(fwId, execId, c, logger)
	e := events.NewSprintExecutorEventController(ex, make(chan *mesos_v1_executor.Event), logger)
	e.Run()
}

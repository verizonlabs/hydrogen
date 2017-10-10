# Hydrogen

## Overview

High performance, fault tolerant Mesos framework based on the v1 streaming API.

Current feature set:
- Launch, destroy, update, and gather state of containers
- Utilizes the universal container runtime instead of Docker
- Configure and use any N number of CNI networks
- High availability of the framework when launching more than one instance
- Filter/constraint support to launch on a set of defined nodes
- UNIQUE or MUX (colocate) deployment strategies
- Custom executor support
- Docker volume support (e.g. rexray, pxd...)

### API Documentation ###
Base endpoint:
<pre><code>http://server:port/v1/api/</code></pre>

#### Examples ####
This is _not_ valid JSON to launch a task but is an example enumeration of all options available.

<pre><code>
[{
  "name": "Example-app",                    # Application Name.
  "resources": {
    "cpu": 1.5,                             # CPU shares
    "mem": 128.25,                          # Memory
    "disk": {"size": 1024.00}               # Disk
  },
  "filters": [                              # Used to filter mesos attributes
      {
        "type": "TEXT",                     # TEXT, SET, SCALAR, RANGES, STRATEGY
        "value": [                          # Example here is filtering on a MAC
          "DEADBEEF00"
        ]
      }
  ],
  "instances": 1,                           # Number of instances to run.
  "strategy": {"type": "unique"},           # Deployment strategy can be set to UNIQUE or MUX.
  "command": {
    "cmd": "/bin/echo hello world",         # Command to run.
    "environment": {
      "VARIABLE_NAME1": "VARIABLE_VALUE1",
      "VARIABLE_NAME2": "VARIABLE_VALUE2"
    },
    "uris": [                               # URI is used to grab the custom executor binary.
      { 
        "uri": "http://some-mesos-dns.mesos:9001/executor",
        "extract": false,                   # Should we extract this file when mesos pulls it?
        "execute": true                     # Should we set the executable bit?
      }
    ]
  },
  "container": {                            # Specifiy a container to launch.
    "image": "debian:latest",               # Which container image to use?
    "volume": [                             # Specifiy any N volumes.
      {
        "container_path": "example/log",    # Specify a static volume path.
        "host_path": "/var/log",
        "mode": "ro"
      },
      {                                     # Docker volume specification
        "source": {
            "type": "DOCKER",               # What type of volume is this?
            "docker_volume": { 
              "driver": "rexray",           # Docker volume driver to use
              "name": "test_volume",        # Docker volume name.
              "driveropts": [""]            # Any driver options you might want to pass.
        }
      }
    ],
    "network": [{                           # Network specification.
      "ipaddress": [{                       # Static IPv4
        "group": ["prod"],                  # Specify a group to belong to.
        "labels": [{"some_label": "awesome"}],
        "ip": "10.2.1.1",
        "protocol": "ipv4"
      },
      {                    
        "group": ["QA"],                    # This interface is also in group "prod".
        "labels": [{"some_label": "rad"}],   # And Add labels.
        "ip": "10.2.1.25",
        "protocol": "ipv4"
      },
      {
        "ip": "2600::1",                    # Static IPv6 address.
        "protocol": "ipv6"
      },
      {
        "name": "FE-CNI"                    # We can specifiy CNI networks by name.
      },
      {
        "name": "BE-CNI"                    # We can have more than one CNI network.
        "labels": [
          {"testIP": "11.11.11.1"},         # Supports passing labels to CNI plugins.
          {"testValue": "someString"}
        ]
      }]
    }]
  },
  "healthcheck": {
    "endpoint": "localhost:8080"            # What endpoint to hit for healthchecks
  },
  "labels": {
    "purpose": "Testing"                    # Labels are supported at the task level as well.
  }
}]
</code></pre>

Example of a valid task that launches onto the CNI network "internal-network"
 and sleeps for 500 seconds.  It uses 0.1 CPU's, 32MB of RAM and 1024MB of disk.
 It will run 5 instances and will launch on unique agents.
<pre><code>
[{
  "name": "test-app",
  "resources": {
    "cpu": 0.1,
    "mem": 32.0,
    "disk": {"size": 1024.00}
  },
  "instances": 5,
  "command": {
    "cmd": "/bin/sleep 500"
  },
  "container": {
    "network": [
     {"name": "internal-network"}
    ]
  },
  "strategy": {"type": "unique"}
}]
</code></pre>

#### Deploy ####
Deploy an application.
<pre><code>Method: POST
/app

# Example
curl -X POST hydrogen.mesos:8080/v1/api/app -d@my-app.json
</pre></code>

#### Kill ####
Kill an application.
<pre><code>Method: DELETE
/app

# Example
curl -X DELETE hydrogen.mesos:8080/v1/api/app -d'{"name": "test-app"}'
</pre></code>

#### Update ####
Update an application.
<pre><code>Method: PUT
/app

# Example
curl -X PUT hydrogen.mesos:8080/v1/api/app -d@my-updated-app.json
</pre></code>

#### State ####
Get the state of an application.
<pre><code>Method: GET
/app

# Example
curl -X GET hydrogen.mesos:8080/v1/api/app?name=test-app
</pre></code>

#### Get All Tasks ####
Get all tasks known to the scheduler
<pre><code>Method: GET
/app/all

# Example
curl -X GET hydrogen.mesos:8080/v1/api/app/all
</pre></code>

### Building ###

#### Requirements ####
Go 1.6 and up.

- `go get -d github.com/verizonlabs/hydrogen/...`
- `cd $GOPATH/src/github.com/verizonlabs/hydrogen && make build`

#### Testing ####

Tests can be run with `make test`. Similarly, you can run benchmarks with `make bench`.

### [License](LICENSE) ###

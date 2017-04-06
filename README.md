# Sprint

## Overview

High performance Mesos framework based on the v1 streaming API.

Current feature set:
- Launch, Destroy, Update, and gather state of a container.
- UCR by default.
- Configure and use any n number of CNI networks.
- High availability of the framework by default.
- Filter/Constraint support to launch on a set of defined nodes.
- Custom executor support.
- Docker volume support (e.g. rexray, pxd...)

Upcoming Features:
(TBD)

## Epics
https://jira.verizon.com/browse/BLCS-138

### API Documentation ###
Base endpoint:
<pre><code>http://server:port/v1/api/</code></pre>

#### Examples ####
This is not valid JSON to launch but an example enumeration of all options available.

<pre><code>
{
  "name": "Example-app",                 # Application Name.
  "resources": { 
    "cpu": 1.5,                          # Use float values
    "mem": 128.25
  },
  "filters": [                           # Used to filter mesos attributes
      {
        "type": "TEXT",                  # TEXT, SET, SCALAR, RANGES
        "value": [                       # Example here is filtering on a MAC
          "DEADBEEF00"                   #
        ]
      },
  ],
  "command": {
    "cmd": "/bin/echo hello world",      # Command to run.
    "uris": [                            # URI is used to grab the custom executor binary.
      { 
        "uri": "http://some-mesos-dns.mesos:9001/executor",
        "extract": false,                # Should we extract this file when mesos pulls it?
        "execute": true                  # Should we set the executable bit?
      }
    ]
  },
  "container": {                         # Specifiy a container to launch.
    "image": "debian:latest",            # Which container image to use?
    "volume": [                          # Specifiy any N volumes.
      {
        "container_path": "example/log", # Specify a static volume path.
        "host_path": "/var/log",
        "mode": "ro"
      },
      {                                  # Docker volume specification
        "source": {
            "type": "DOCKER",            # What type of volume is this?
            "docker_volume": { 
              "driver": "rexray",        # Docker volume driver to use
              "name": "test_volume",     # Docker volume name.
              "driveropts": [""]         # Any driver options you might want to pass.
        }
      }
    ],
    "network": [{                        # Network specification.
      "ipaddress": [{                    # Static IPv4
        "group": ["prod"],               # Specify a group to belong to.
        "label": [{"some_label": "awesome"}],
        "ip": "10.2.1.1",
        "protocol": "ipv4"
      },
      {                    
        "group": ["QA"],                 # This interface is also in group "prod".
        "label": [{"some_label": "rad"}],# And Add labels.
        "ip": "10.2.1.25",
        "protocol": "ipv4"
      },
      {
        "ip": "2600::1",                 # Static IPv6 address.
        "protocol": "ipv6"
      },
      {
        "name": "FE-CNI"                 # We can specifiy CNI networks by name.
      },
      {
        "name": "BE-CNI"                 # We can have more than one CNI network...
      }]
    }]
  },
  "healthcheck": {
    "endpoint": "localhost:8080"         # What endpoint to hit for healthchecks
  },
  "labels": [{
    "purpose": "Testing"                 # Labels are supported.
  }]
}
</code></pre>

#### Deploy ####
Deploy an application.
<pre><code>Method: POST
/deploy
</pre></code>

#### Kill ####
Kill an application.
<pre><code>Method: DELETE
/kill
</pre></code>

#### Update ####
Update an application.
<pre><code>Method: PUT
/update
</pre></code>

#### State ####
Get the state of an application.
<pre><code>Method: GET
/state
</pre></code>
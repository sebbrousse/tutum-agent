tutum-agent
===========

Agent to control Tutum nodes


# Functions

* Download docker binary
* Register new nodes with Tutum
* Launch docker daemon
* Auto restart docker daemon on failure

## Run Tutum-agent 
```
# ./tutum-agent -h     
Usage of tutum-agent:
  -debug=false: Enable debug mode
  -stdout=false: Print log to stdout
   set: Set items in the config file and exit, supported items
          CertCommonName="xxx"
          DockerBinaryURL="xxx"
          DockerHost="xxx"
          TutumHost="xxx"
          TutumToken="xxx"
          TutumUUID="xxx"
```


Configruation file is put in `/etc/tutum/agent/tutum-agent.conf` (json file)

Items in `tutum-agent.conf`:
```
{
	"CertCommonName":"*.node.tutum.io",
	"DockerBinaryURL":"https://files.tutum.co/packages/docker/latest.json",
	"DockerHost":"tcp://0.0.0.0:2375",
	"TutumHost":"https://dashboard.tutum.co/",
	"TutumToken":"token",
	"TutumUUID":"uuid"
}

```

Const var used in  `tutum-agent`
```
const (
        defaultCertCommonName  = ""
        defaultDockerHost      = "tcp://0.0.0.0:2375"
        defaultDockerBinaryURL = "https://files.tutum.co/packages/docker/latest.json"
        defaultTutumHost       = "https://dashboard.tutum.co/"
)

const (
        TutumHome = "/etc/tutum/agent"
        DockerDir = "/usr/lib/tutum"
        LogDir    = "/var/log/tutum"

        DockerLogFile   = "docker.log"
        TutumLogFile    = "agent.log"
        KeyFile         = "key.pem"
        CertFile        = "cert.pem"
        CAFile          = "ca.pem"
        ConfigFile      = "tutum-agent.conf"
        DockerBinary    = "docker"
        DockerNewBinary = "docker.new"

        RegEndpoint       = "api/agent/node/"
        DockerDefaultHost = "unix:///var/run/docker.sock"

        MaxWaitingTime    = 200 //seconds
        HeartBeatInterval = 5   //second
)

```

## Building

Run

	make

to build binaries and `.deb` packages which will be copied to the `build/` folder.


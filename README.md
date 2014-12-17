tutum-agent
===========


## What's this?

This is the agent Tutum uses to set up nodes. It's a daemon that will register the host with the Tutum API using a user token (`TutumToken`), and will manage the installation, configuration and ongoing upgrade of the Docker daemon.

For information on how to install it in your host, please check the [Bring Your Own Node](https://support.tutum.co/support/solutions/articles/5000513678-bring-your-own-node) documentation.


## Running

If installing from the `.deb` package, Tutum Agent will be configured in upstart to be launched on boot.

```
# tutum-agent -h     
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


Configuration file is located in `/etc/tutum/agent/tutum-agent.conf` (JSON file) with the following structure:

```
{
	"CertCommonName":"*.node.tutum.io",
	"DockerBinaryURL":"https://files.tutum.co/packages/docker/latest.json",
	"DockerHost":"tcp://0.0.0.0:2375",
	"TutumHost":"https://dashboard.tutum.co/",
	"TutumToken":"<token>",
	"TutumUUID":"<uuid>"
}
```

## Logging

Logs are stored under `/var/log/tutum/`:

* `agent.log` contains the logs of the agent itself
* `docker.log` contains the Docker daemon logs


## Building

Run `make` to build binaries and `.deb` packages which will be stored in the `build/` folder.


## Known limitations

Currently only tested on Ubuntu 14.04


## Reporting security issues

In order to report a security issue, please send us an email to [security@tutum.co](mailto:security@tutum.co). Please use GPG key ID `666DAA4A` on `keys.gnupg.net` to encrypt your email. Thank you!



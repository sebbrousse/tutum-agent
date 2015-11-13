#!/bin/sh
#
# Usage:
# curl -Ls https://get.tutum.co/ | sudo -H sh -s [TutumToken] [TutumUUID] [CertCommonName]
#
set -e
GPG_KEY_TUTUM_ID=A87A2270
GPG_KEY_PACKAGE_ID=${GPG_KEY_PACKAGE_ID:-A87A2270}
GPG_KEY_TUTUM_URL=https://files.tutum.co/keys/$GPG_KEY_TUTUM_ID.pub
GPG_KEY_PACKAGE_URL=https://files.tutum.co/keys/$GPG_KEY_PACKAGE_ID.pub
TUTUM_REPO=${TUTUM_REPO:-repo.tutum.co}
TUTUM_HOST=${TUTUM_HOST:-https://dashboard.tutum.co/}
SUPPORT_URL=http://go.tutum.co/support-byon
export DEBIAN_FRONTEND=noninteractive


if [ -f "/etc/tutum/agent/tutum-agent.conf" ]; then
	cat <<EOF
ERROR: Tutum Agent is already installed in this host
If the node failed to register properly with Tutum, try to restart the agent by executing:

	service tutum-agent restart

and check the logs at /var/log/tutum/agent.log for possible errors.
If the problem persists, please contact us at support@tutum.co
EOF
	exit 1
fi

if [ "$(uname -m)" != "x86_64" ]; then
	cat <<EOF
ERROR: Unsupported architecture: $(uname -m)
Only x86_64 architectures are supported at this time
Learn more: $SUPPORT_URL
EOF
	exit 1
fi

get_distribution_type()
{
	local lsb_dist
	lsb_dist="$(lsb_release -si 2> /dev/null || echo "unknown")"
	if [ "$lsb_dist" = "unknown" ]; then
		if [ -r /etc/lsb-release ]; then
			lsb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
		elif [ -r /etc/debian_version ]; then
			lsb_dist='debian'
		elif [ -r /etc/fedora-release ]; then
			lsb_dist='fedora'
		elif [ -r /etc/centos-release ]; then
			lsb_dist='centos'
		elif [ -r /etc/redhat-release ]; then
			lsb_dist='rhel'
		elif [ -r /etc/os-release ]; then
			lsb_dist="$(. /etc/os-release && echo "$ID")"
		fi
	fi
	lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"
	echo $lsb_dist
}

case "$(get_distribution_type)" in
	ubuntu|debian)
		echo "-> Adding Tutum's GPG key..."
		curl -Ls --retry 30 --retry-delay 10 $GPG_KEY_TUTUM_URL | gpg --import -
		curl -Ls --retry 30 --retry-delay 10 $GPG_KEY_PACKAGE_URL | apt-key add -
		echo "-> Installing required dependencies..."
		modprobe -q aufs || apt-get update -qq && apt-get install -yq linux-image-extra-$(uname -r) || \
			echo "!! Failed to install linux-image-extra package. AUFS support (which is recommended) may not be available."
		echo "-> Installing tutum-agent..."
		echo deb [arch=amd64] http://$TUTUM_REPO/ubuntu/ tutum main > /etc/apt/sources.list.d/tutum.list
		apt-get update -qq -o Dir::Etc::sourceparts="/dev/null" -o APT::List-Cleanup=0 -o Dir::Etc::sourcelist="sources.list.d/tutum.list" && apt-get install -yq tutum-agent
		;;
	fedora|centos|rhel)
		echo "-> Adding Tutum's GPG key..."
		yum install -y gpg rpm curl
		curl -Ls --retry 30 --retry-delay 10 $GPG_KEY_TUTUM_URL | gpg --import -
		rpm --import $GPG_KEY_PACKAGE_URL
		echo "-> Installing tutum-agent..."
		cat > /etc/yum.repos.d/tutum.repo <<EOF
[tutum]
name=Tutum
baseurl=http://$TUTUM_REPO/redhat/\$basearch
enabled=1
gpgkey=$GPG_KEY_PACKAGE_URL
repo_gpgcheck=1
gpgcheck=1
EOF
		yum install -y tutum-agent
		;;
	*)
		echo "ERROR: Cannot detect Linux distribution or it's unsupported"
		echo "Learn more: $SUPPORT_URL"
		exit 1
		;;
esac

echo "-> Configuring tutum-agent..."
mkdir -p /etc/tutum/agent
cat > /etc/tutum/agent/tutum-agent.conf <<EOF
{
	"TutumHost":"${TUTUM_HOST}",
	"TutumToken":"${1}",
	"TutumUUID":"${2}",
	"CertCommonName":"${3}"
}
EOF

if [ -d /run/systemd/system ] ; then
	echo "-> Enabling tutum-agent to start on boot on systemd..."
	systemctl enable tutum-agent.service || true
fi

if [ ! -z "${1}" ]; then
	echo "-> Starting tutum-agent service..."
	service tutum-agent stop > /dev/null 2>&1 || true
	service tutum-agent start
fi

echo "-> Done!"
cat <<EOF

*******************************************************************************
Tutum Agent installed successfully
*******************************************************************************

You can now deploy containers to this node using Tutum

EOF

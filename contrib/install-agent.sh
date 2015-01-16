#!/bin/sh
#
# Usage:
# curl -Ls https://files.tutum.co/scripts/install-agent.sh | sudo -H sh -s [TutumToken] [TutumUUID] [CertCommonName]
#
set -e
GPG_KEY_ID=A87A2270
S3_BUCKET=repo.tutum.co
TUTUM_HOST=https://dashboard.tutum.co/
export DEBIAN_FRONTEND=noninteractive

echo "-> Adding Tutum's GPG key..."
apt-key adv --keyserver hkp://keyserver.ubuntu.com --recv-keys $GPG_KEY_ID > /dev/null
gpg --keyserver hkp://keyserver.ubuntu.com --recv-keys $GPG_KEY_ID > /dev/null
echo "-> Installing required dependencies..."
apt-get update -qq && apt-get install -yq apt-transport-https curl > /dev/null
apt-get install -yq linux-image-extra-$(uname -r) > /dev/null || \
    echo "Failed to install linux-image-extra package. AUFS support may not be available."
echo "-> Installing tutum-agent..."
echo deb [arch=amd64] https://$S3_BUCKET/ubuntu/ tutum main > /etc/apt/sources.list.d/tutum.list
apt-get update -qq && apt-get install -yq tutum-agent > /dev/null
echo "-> Configuring tutum-agent..."
mkdir -p /etc/tutum/agent
tutum-agent set TutumHost="$TUTUM_HOST" > /dev/null
if [ $# -gt 0 ]; then
	tutum-agent set TutumToken="$1" > /dev/null
fi
if [ $# -gt 1 ]; then
	tutum-agent set TutumUUID="$2" > /dev/null
fi
if [ $# -gt 2 ]; then
	tutum-agent set CertCommonName="$3" > /dev/null
fi
service tutum-agent restart > /dev/null 2>&1
echo "-> Done!"
cat <<EOF

*******************************************************************************
Tutum Agent installed successfully
*******************************************************************************

You can now deploy containers to this node using Tutum

EOF

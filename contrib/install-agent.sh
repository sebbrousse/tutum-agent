#!/bin/sh
#
# Usage:
# curl -Ls https://get.tutum.co/ | sudo -H sh -s [TutumToken] [TutumUUID] [CertCommonName]
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
modprobe -q aufs || apt-get update -qq && apt-get install -yq linux-image-extra-$(uname -r) || \
    echo "!! Failed to install linux-image-extra package. AUFS support may not be available."
echo "-> Installing tutum-agent..."
echo deb [arch=amd64] http://$S3_BUCKET/ubuntu/ tutum main > /etc/apt/sources.list.d/tutum.list
apt-get update -qq -o Dir::Etc::sourceparts="/dev/null" -o APT::List-Cleanup=0 -o Dir::Etc::sourcelist="sources.list.d/tutum.list" && apt-get install -yq tutum-agent
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
service tutum-agent restart > /dev/null 2>&1
echo "-> Done!"
cat <<EOF

*******************************************************************************
Tutum Agent installed successfully
*******************************************************************************

You can now deploy containers to this node using Tutum

EOF

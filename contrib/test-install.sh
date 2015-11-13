#!/bin/sh

export TUTUM_REPO=repo-test.tutum.co.s3.amazonaws.com
export GPG_KEY_PACKAGE_ID=90E64D7C
export TUTUM_HOST="https://dashboard-staging.tutum.co"

if [ -z "$1" ]; then
	echo "token is not provided"
	exit 1
fi

if which sudo >/dev/null 2>&1; then
	#curl -Ls https://get.tutum.co/ | sudo -H sh -s $1
	cat install-agent.sh | sudo -H TUTUM_REPO=${TUTUM_REPO} GPG_KEY_PACKAGE_ID=${GPG_KEY_PACKAGE_ID} TUTUM_HOST=${TUTUM_HOST} sh -s $1
else
	curl -Ls https://get.tutum.co/ | TUTUM_REPO=${TUTUM_REPO} GPG_KEY_PACKAGE_ID=${GPG_KEY_PACKAGE_ID} TUTUM_HOST=${TUTUM_HOST} sh -s $1
fi

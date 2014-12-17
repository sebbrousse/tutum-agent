#!/bin/bash
set -e

if [ ! -f /package.deb ]; then
	echo "No /package.deb found"
	exit 1
fi

gpg --import /gpg_public_key
gpg --import /gpg_private_key
mkdir -p /repo/db/
mkdir -p /repo/dists/
aws s3 sync s3://$S3_BUCKET/ubuntu/db/ /repo/db/ --region us-east-1
aws s3 sync s3://$S3_BUCKET/ubuntu/dists/ /repo/dists/ --region us-east-1
reprepro --keepunusednewfiles --ask-passphrase -Vb /repo includedeb tutum /package.deb
aws s3 sync /repo/ s3://$S3_BUCKET/ubuntu/ --acl public-read --region us-east-1

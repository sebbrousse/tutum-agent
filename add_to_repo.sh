#!/bin/bash
set -e
# Production
# GPG_KEY_ID=A87A2270
# S3_BUCKET=repo.tutum.co

# Staging
GPG_KEY_ID=90E64D7C
S3_BUCKET=repo-test.tutum.co

if [ ! -f "$1" ]; then
	echo "Invalid deb package $1"
	exit 1
fi

cd repo/
cp ../$1 ./package.deb
gpg --send-keys --keyserver keyserver.ubuntu.com $GPG_KEY_ID
gpg --export -a $GPG_KEY_ID > ./gpg_public_key
gpg --export-secret-key -a $GPG_KEY_ID > ./gpg_private_key
docker build -t agentrepo .
docker run --rm -i -t -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e S3_BUCKET=$S3_BUCKET agentrepo
docker rmi agentrepo
rm -f ./package.deb

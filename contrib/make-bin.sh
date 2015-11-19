#!/bin/bash

VERSION=$(cat VERSION)
DEST=$1
GOARCH=arm
GOOS=${GOOS:-linux} GOARCH=${GOARCH:-amd64} go build -o ${DEST:-/build/bin}/${GOOS:-linux}/${GOARCH:-amd64}/tutum-agent-${VERSION:-latest}
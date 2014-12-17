#!/bin/bash

mkdir -p /build/{bin,ubuntu}
contrib/make-bin.sh /build/bin
contrib/make-deb.sh /build/ubuntu
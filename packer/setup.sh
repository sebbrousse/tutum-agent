#!/bin/bash
set -e
mkdir -p staging
rm -f staging/*.box
wget https://vagrantcloud.com/ubuntu/boxes/trusty64/versions/14.04/providers/virtualbox.box -P staging
tar zxvf staging/*.box -C staging/
rm -f staging/*.box

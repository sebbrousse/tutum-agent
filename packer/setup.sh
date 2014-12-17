#!/bin/bash
mkdir staging
wget https://vagrantcloud.com/ubuntu/trusty64/version/1/provider/virtualbox.box -P staging
tar zxvf staging/*.box -C staging/
rm -f staging/*.box
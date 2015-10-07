#!/bin/sh

cp /scripts/upgrade-tutum-agent.sh /rootfs/tmp/upgrade-tutum-agent.sh 
chroot /rootfs /tmp/upgrade-tutum-agent.sh

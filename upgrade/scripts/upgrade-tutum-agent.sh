#!/bin/bash

AGENT_PIDFILE="/var/run/tutum-agent.pid"
AGENT_BINARY=$(which tutum-agent)


get_distribution_type()
{
    local lsb_dist
    lsb_dist="$(lsb_release -si 2> /dev/null || echo "unknown")"
    if [ "$lsb_dist" = "unknown" ]; then
        if [ -r /etc/lsb-release ]; then
            lsb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
        elif [ -r /etc/debian_version ]; then
            lsb_dist='debian'
        elif [ -r /etc/fedora-release ]; then
            lsb_dist='fedora'
        elif [ -r /etc/centos-release ]; then
            lsb_dist='centos'
        elif [ -r /etc/os-release ]; then
            lsb_dist="$(. /etc/os-release && echo "$ID")"
        fi
    fi
    lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"
    echo $lsb_dist
}

get_agent_pid(){
    if [ -f ${AGENT_PIDFILE} ]; then
        cat /var/run/tutum-agent.pid
    else
        ps ax | grep -F tutum-agent | grep -v grep | grep -v sudo | head -n 1 | cut -d " " -f 1 2>/dev/null
    fi
}


get_agent_version()
{
    ver=$(${AGENT_BINARY} -v 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo ${ver}
    else
        echo "unknow version, below 0.18.1"
    fi
    unset ver
}

upgrade_on_ubuntu()
{
    apt-get update || true
    apt-get install -y tutum-agent
}


OLD_AGENT_VERSION=$(get_agent_version)
AGENT_PID=$(get_agent_pid)
if [ -n "${AGENT_PID}" ]; then
    echo "=> tutum-agent(${OLD_AGENT_VERSION}) is running with PID: ${AGENT_PID}"
else
    echo "=> tutum-agent(${OLD_AGENT_VERSION}) is running with PID: unknown"
fi

case "$(get_distribution_type)" in
    ubuntu)
        echo "=> host operating system detected: ubuntu"
        upgrade_on_ubuntu
        ;;
    *)
        echo "=> error: Cannot detect Linux distribution or it's unsupported"
        exit 1
        ;;
esac

NEW_AGENT_VERSION=$(get_agent_version)
if [ "${OLD_AGENT_VERSION}" == "${NEW_AGENT_VERSION}" ]; then
    echo "=> version of tutum-agent remains the same"
    echo "=> exiting without any changes"
else
    echo "=> tutum-agent is upgraded from ${OLD_AGENT_VERSION} to ${NEW_AGENT_VERSION}"
    if [ -n "${AGENT_PID}" ]; then
        echo "=> killing the current tutum-agent process, and it will be restarted by upstart/systemd/sysmvinit"
        echo "=> NOTICE: you might have to restart your stopped containers if they are launched without autorestart option"
        kill ${AGENT_PID}
    else
        echo "=> Please restart tutum-agent to apply the changes"
    fi
fi

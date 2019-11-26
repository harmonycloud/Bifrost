#!/bin/sh

# Script to build  linux bridge bro for hcipam CNI on a Kubernetes host.

set -u -e

set -x

device_name=$1

echo  'get device name '$device_name

check_str=`ip addr show $device_name|grep $device_name|grep mtu|awk '{print $4}'`

echo 'check str is '$check_str

if [ $check_str = "mtu" ]; then
    echo 'check iproute and device true'
else
    echo 'check iproute and device false'
    exit 1
fi

deviceIP=`ip addr list $device_name|grep -v inet6|grep inet|awk '{print $2}'`

deviceMac=`ip link show $device_name|grep ether|awk '{print $2}'`

defaultRoute=`ip route | grep default |grep $device_name| awk '{print $3}'`

br0Ex=`ip a|grep br0|grep state|awk '{print $2}'`

function config_br0() {

    if [ -z "$br0Ex" ]; then
          echo 'br0 not exist,create new'
         `ip link add name br0 type bridge`
    else
           echo "br0 exist"
    fi

    `ip link set br0 up`
    if [ $? -eq 0 ]; then
        echo "set bridge br0 up... yes"
    else
        echo "set bridge br0 up.. no"
        exit 1
    fi
    `ip addr add $deviceIP dev br0`

    if [ $? -eq 0 ]; then
        echo "set br0 deviceIP... yes"
    else
        echo "set br0 deviceIP... no"
        exit 1
    fi
    `ip addr delete $deviceIP dev $device_name`

    if [ $? -eq 0 ]; then
        echo "delete deviceIP... yes"
    else
        echo "delete deviceIP... no"
        exit 1
    fi

    #正确执行完后网络恢复
    `ip route add default via $defaultRoute`

    if [ $? -eq 0 ]; then
        echo "set br0 default route... yes"
    else
        echo "set br0 default route ... no"
        exit 1
    fi

}

config_br0

#执行该命令断网
`ip link set $device_name master br0`

if [ $? -eq 0 ]; then
    echo "set master... yes"
else
    echo "set master... no"
	exit 1
fi

   `ip link set br0 addr $deviceMac`
    if [ $? -eq 0 ]; then
        echo "set br0  mac... yes"
    else
        echo "set br0 mac ... no"
        exit 1
    fi

`modprobe br_netfilter`

if [ $? -eq 0 ]; then
    echo "modprobe br_netfilter... yes"
else
    echo "modprobe br_netfilter ... no"
	exit 1
fi

echo `sysctl -w net.ipv4.ip_forward\=1`

echo `sysctl -w net.bridge.bridge-nf-call-iptables\=1`

echo `ip link set br0 promisc on`

echo `iptables -P FORWARD ACCEPT`


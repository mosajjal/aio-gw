#!/bin/sh
set -e

# run like this:
# /polarproxy-entrypoint.sh [polarproxy args]

ip tuntap add mode tun dev tun0 && \
ip addr add 100.100.100.100/32 dev tun0 && \
ip link set dev tun0 up
# TODO: egress is a hardcoded container name that needs to be an env variable
opt/tun2socks/tun2socks-linux-amd64 -device tun0 -proxy socks5://egress:1080 &
# remove default route and replace with tun0
ip route del default
ip route add default dev tun0
sleep 1

cd /opt/polarproxy
dotnet PolarProxy.dll $@
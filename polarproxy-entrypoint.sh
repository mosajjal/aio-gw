#!/bin/sh
set -e

# run like this:
# /polarproxy-entrypoint.sh [polarproxy args]

ip tuntap add mode tun dev tun0 && \
ip addr add 100.100.100.100/32 dev tun0 && \
ip link set dev tun0 up
# TODO: egress is a HARDCODED container name that needs to be an env variable
opt/tun2socks/tun2socks-linux-amd64 -device tun0 -proxy socks5://egress:1080 &
# remove default route and replace with tun0

sh -c "sleep 30 && ip route del default && ip route add default dev tun0 && sleep 1" &

# run dump2poip to convert dump to poip
# TODO: arkime is HARDCODED
/opt/polarproxy/dump2poip -be -i tun0 -o tcp://arkime:57012 &

cd /opt/polarproxy
dotnet PolarProxy.dll $@
#!/bin/sh
# kill parent processes
cleanup() {
    for i in $(seq 0 3); do kill $(cat test/run/qemu-"$i".pid); done
    pkill -P $$
}

# Setup signals to kill child processes on exit.
for sig in INT QUIT HUP TERM; do
  trap "
    cleanup
    trap - $sig EXIT
    kill -s $sig "'"$$"' "$sig"
done
trap cleanup EXIT

timeout="120"
bridge="br0"

talos_version="1.0.5"
registry_version="2.8.1"
talos_arch="amd64"

if type lsb_release >/dev/null 2>&1 ; then
   DISTRO=$(lsb_release -i -s)
elif [ -e /etc/os-release ] ; then
   DISTRO=$(awk -F= '$1 == "ID" {print $2}' /etc/os-release)
fi

DISTRO=$(printf '%s\n' "$DISTRO" | LC_ALL=C tr '[:upper:]' '[:lower:]')

case "$DISTRO" in
    nixos*) PATH="/run/current-system/sw/bin:$PATH" mount --bind test/etc/resolv.conf /etc/resolv.conf ;;
    *)      mount --bind test/etc/resolv.conf /etc/resolv.conf ;;
esac

tapup() {
    ip tuntap add dev "$1" mode tap
    ip link set dev "$1" master "$bridge"
    ip link set "$1" up
}

ip link add br0 type bridge
ip link set dev tap9 master br0
ip link set dev br0 up

ip address delete 10.0.2.100/24 dev tap9
ip address add 10.0.2.100/24 dev br0
ip route add default via 10.0.2.2 dev br0

sleep 3

# Predefined MAC addresses
MACS="de:ad:be:ef:54:be de:ad:be:ef:ec:72 de:ad:be:ef:88:c0 de:ad:be:ef:41:1c"

mkdir -p test/run/registry

REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io \
    REGISTRY_HTTP_ADDR=0.0.0.0:5000 \
    REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/docker.io \
    test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/registry-1.docker.io.log 2>&1 &

REGISTRY_PROXY_REMOTEURL=https://k8s.gcr.io \
    REGISTRY_HTTP_ADDR=0.0.0.0:5001 \
    REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/k8s.gcr.io \
    test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/k8s.gcr.log 2>&1 &

REGISTRY_PROXY_REMOTEURL=https://quay.io \
    REGISTRY_HTTP_ADDR=0.0.0.0:5002 \
    REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/quay.io \
    test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/quay.io.log 2>&1 &

REGISTRY_PROXY_REMOTEURL=https://gcr.io \
    REGISTRY_HTTP_ADDR=0.0.0.0:5003 \
    REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/gcr.io \
    test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/gcr.io.log 2>&1 &

REGISTRY_PROXY_REMOTEURL=https://ghcr.io \
    REGISTRY_HTTP_ADDR=0.0.0.0:5004 \
    REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/ghcr.io \
    test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/ghcr.io.log 2>&1 &

mkdir -p /tmp/qmp

vmup() {
    kvm="--enable-kvm"

    MAC=$(echo $MACS | awk -F ' ' "{print \$($1+1)}")

    qemu-img create -f qcow2 "test/opt/vm-$1.qcow2" 4G
    qemu-system-x86_64 -m 2048 -boot d  \
		       -cdrom test/opt/talos-amd64-v1.0.5.iso \
		       -drive file=test/opt/vm-"$1".qcow2,format=qcow2,if=virtio \
		       -netdev tap,id=mynet0,ifname=tap"$1",script=no,downscript=no \
		       -device virtio-net-pci,netdev=mynet0,mac=$MAC \
		       -serial file:test/run/vm-"$1".log -display none \
		       -daemonize -pidfile test/run/qemu-"$1".pid \
		       -device virtio-rng-pci \
    		       -qmp unix:/tmp/qmp/vm-node-"$1".sock,server,nowait \
		       ${kvm}

    tail -f test/run/vm-"$1".log | sed "s/^/(vm-node-$1 LOG): /" &
}

mkdir -p test/run
# Create VM pool
for i in $(seq 0 3); do
    tapup tap"$i";
    vmup "$i"
done

(
    while socket=$(inotifywait -q /tmp/qmp -e delete --format '%f'); do {
	if [ "${socket: -5}" == ".sock" ]; then
	    echo "restarting $socket"
	    vmup $(echo "$socket" | sed 's/[^0-9]*//g')
	fi
    }
    done
) &

sleep 2

# Run acceptance tests
TF_ACC=1 TALOSCONF_DIR=$(pwd)/test/run go test -v ./talos -timeout 120m

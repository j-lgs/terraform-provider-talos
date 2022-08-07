#!/usr/bin/env bash

# Variables
nNodes=4
nodes=$(($nNodes-1))

timeout="120"
bridge="br0"

talos_version="1.0.5"
registry_version="2.8.1"
talos_arch="amd64"

sentinel="test/run/sentinel"

# Do not tolerate errors
set -e

# Determine the distro
if type lsb_release >/dev/null 2>&1 ; then
   DISTRO=$(lsb_release -i -s)
elif [ -e /etc/os-release ] ; then
   DISTRO=$(awk -F= '$1 == "ID" {print $2}' /etc/os-release)
fi
DISTRO=$(printf '%s\n' "$DISTRO" | LC_ALL=C tr '[:upper:]' '[:lower:]')

# Cleanup script
function finish {
    # unmount resolvconf
    case "$DISTRO" in
	nixos*) PATH="/run/current-system/sw/bin:$PATH" umount /etc/resolv.conf ;;
	*)      umount /etc/resolv.conf ;;
    esac

    # Kill all registry processes
    registry_down

    echo 'kill' >$sentinel

    for pid in ${pids[*]}; do
	pkill -P $pid
    done

    # long cooldown time before exiting, to ensure the processes are dead and stay dead
    sleep 10
    rm -f $sentinel

    echo "exit success"
}
trap finish EXIT

# Set up a tuntap for a VM inside the network namespace
tapup() {
    ip tuntap add dev "$1" mode tap
    ip link set dev "$1" master "$bridge"
    ip link set "$1" up
}

function registry_up() {
    # Registrys will be cached to speed up runs and avoid docker pulls
    mkdir -p test/run/registry

    REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io \
	REGISTRY_HTTP_ADDR=0.0.0.0:5000 \
	REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/docker.io \
	test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/registry-1.docker.io.log 2>&1 &
    REGISTRY_1_PID="$!"

    REGISTRY_PROXY_REMOTEURL=https://k8s.gcr.io \
	REGISTRY_HTTP_ADDR=0.0.0.0:5001 \
	REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/k8s.gcr.io \
	test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/k8s.gcr.log 2>&1 &
    REGISTRY_2_PID="$!"

    REGISTRY_PROXY_REMOTEURL=https://quay.io \
	REGISTRY_HTTP_ADDR=0.0.0.0:5002 \
	REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/quay.io \
	test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/quay.io.log 2>&1 &
    REGISTRY_3_PID="$!"

    REGISTRY_PROXY_REMOTEURL=https://gcr.io \
	REGISTRY_HTTP_ADDR=0.0.0.0:5003 \
	REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/gcr.io \
	test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/gcr.io.log 2>&1 &
    REGISTRY_4_PID="$!"

    REGISTRY_PROXY_REMOTEURL=https://ghcr.io \
	REGISTRY_HTTP_ADDR=0.0.0.0:5004 \
	REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=test/registry/ghcr.io \
	test/bin/registry-v${registry_version} serve test/etc/registry.yml > test/run/ghcr.io.log 2>&1 &
    REGISTRY_5_PID="$!"
}

function registry_down() {
    kill $REGISTRY_1_PID $REGISTRY_2_PID $REGISTRY_3_PID $REGISTRY_4_PID $REGISTRY_5_PID
}

# Predefined MAC addresses
MACS="de:ad:be:ef:54:be de:ad:be:ef:ec:72 de:ad:be:ef:88:c0 de:ad:be:ef:41:1c"

vm_pids=()

# bring a VM up
vmup() {
    kvm="--enable-kvm"

    MAC=$(echo $MACS | awk -F ' ' "{print \$($1+1)}")

    qemu-img create -f qcow2 "test/opt/vm-$1.qcow2" 4G
    qemu-system-x86_64 -m 2048 -boot d  \
		       -cpu host \
		       -cdrom test/opt/talos-amd64-v1.1.0.iso \
		       -drive file=test/opt/vm-"$1".qcow2,format=qcow2,if=virtio \
		       -netdev tap,id=mynet0,ifname=tap"$1",script=no,downscript=no \
		       -device virtio-net-pci,netdev=mynet0,mac=$MAC \
		       -serial stdio \
		       -display none \
		       -device virtio-rng-pci \
    		       -qmp unix:/tmp/qmp/vm-node-"$1".sock,server,nowait \
		       ${kvm} | sed -e "s/^/VM Node $1: /;"
}

# mount resolv.conf for DNS
case "$DISTRO" in
    nixos*) PATH="/run/current-system/sw/bin:$PATH" mount --bind test/etc/resolv.conf /etc/resolv.conf ;;
    *)      mount --bind test/etc/resolv.conf /etc/resolv.conf ;;
esac

# set up a bridge for all VM tap interfaces and set up a gateway IP and route
ip link add br0 type bridge
ip link set dev tap9 master br0
ip link set dev br0 up

ip address delete 10.0.2.100/24 dev tap9
ip address add 10.0.2.100/24 dev br0
ip route add default via 10.0.2.2 dev br0

# wait for network to configure
sleep 3

# set registries up
registry_up

mkdir -p /tmp/qmp

pids=()

rm -f $sentinel
touch $sentinel

# Create VM pool
for i in $(seq 0 $nodes); do
    # Set up VM tuntap
    tapup tap"$i";

    # in the background poll whether the VM needs to be restarted, and restart it when required
    (while ! [ -s $sentinel ]; do
	echo "Qemu VM '$i' (re)starting" >&2
	vmup "$i" && break
	sleep 1
    done)&
    pids[$i]="$!"
done

# wait for VMs to initialise
sleep 2

# Run acceptance tests
TF_ACC=1 TALOSCONF_DIR=$(pwd)/test/run go test -v ./talos -timeout 120m
exit 0

#!/bin/sh
docker stop gcr_mirror k8s_gcr_mirror ghcr_mirror docker_mirror quay_mirror
docker rm gcr_mirror k8s_gcr_mirror ghcr_mirror docker_mirror quay_mirror
virsh -c qemu:///system destroy test_control_0
virsh -c qemu:///system destroy test_control_1
virsh -c qemu:///system destroy test_control_2
virsh -c qemu:///system destroy test_control_3
virsh -c qemu:///system undefine test_control_0 --remove-all-storage
virsh -c qemu:///system undefine test_control_1 --remove-all-storage
virsh -c qemu:///system undefine test_control_2 --remove-all-storage
virsh -c qemu:///system undefine test_control_3 --remove-all-storage
virsh -c qemu:///system net-destroy talos-acctest
virsh -c qemu:///system net-undefine talos-acctest


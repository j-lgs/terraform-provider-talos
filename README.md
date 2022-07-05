[![Go](https://github.com/j-lgs/terraform-provider-talos/actions/workflows/go.yml/badge.svg)](https://github.com/j-lgs/terraform-provider-talos/actions/workflows/go.yml)

# provider-terraform-talos
A terraform provider for the [Talos Kubernetes OS](https://github.com/siderolabs/talos) from Siderolabs. Exposes worker nodes, controlplane nodes, and the cluster's configuration as terraform resources.

So far this provider isn't ready for most people's use, as it's a personal project to learn the go programming language. The version number will reflect this fact until it reaches an acceptable quality.

Check it out [Here](https://registry.terraform.io/providers/j-lgs/talos/latest)

# Use
Check the examples folder to see how the provider can be used. Also check out my [homelab provisioning](https://github.com/j-lgs/provisioning) repo to see the provider used to set up a Kubernetes cluster on Proxmox hosts.

The current provider version targets Talos v1.1.0.

This project is developed on an AMD64 Linux platform. Releases have not been tested on alternative platforms yet.

## Limitations
So far
+ Node resource deletion doesn't take into account cluster membership
+ Worker nodes don't work properly yet
+ Read, Update and Deletion are not properly implemented

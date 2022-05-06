# provider-terraform-talos
A terraform provider for the [Talos Kubernetes OS](https://github.com/siderolabs/talos) from Siderolabs. Exposes worker nodes, controlplane nodes, and the cluster's configuration as terraform resources.

So far this provider isn't ready for most people's use, as it's a personal project to learn the go programming language. The version number will reflect this fact until it reaches an acceptable quality.

# Use
Check the examples folder to see how the provider can be used. Also check out my [homelab provisioning](https://github.com/j-lgs/provisioning) repo to see the provider used to set up a Kubernetes cluster on Proxmox hosts.

## Limitations
So far
+ Multiple interfaces cannot be specified for a node
+ Multiple wireguard links cannot be specified for a node
+ Node resource deletion doesn't take into account cluster membership
+ Worker nodes don't work properly yet
+ The full range of config options are not covered by either resource

## Known issues
+ Resource input values are not checked
+ There are no acceptance or unit tests

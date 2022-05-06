terraform {
  required_providers {
    talos = {
      source  = "localhost/jlgs/talos"
      version = "0.0.1"
    }
  }
}

resource "talos_configuration" "single_example" {
  # Talos configuration version target
  target_version = "v1.0"
  # Name of the talos cluster
  cluster_name = "taloscluster"
  # List of control plane nodes to act as endpoints the talos client should connect to
  endpoints = ["10.0.1.100"]

  # The evential endpoint that the kubernetes client will connect to
  kubernetes_endpoint = "https://10.0.1.100:6443"

  # The kubernetes version to be deployed
  kubernetes_version = "1.23.6"
}

resource "talos_control_node" "single_example" {
  # The node's hostname
  name = "cluster-control-1"

  # MAC address for the node. Will be used to apply the initial configuration
  macaddr = "0a:0a:0a:42:42:42"
  # Expected network range to find the node's interface identified by `macaddr`
  dhcp_network_cidr = "10.0.0.0/24"
  # The disk to install talos onto
  install_disk = "/dev/vda"

  # The primary interface's address in CIDR form
  ip = "10.0.1.100/17"
  # The primary interface's gateway address
  gateway = "10.0.0.1"
  # The node's nameservers
  nameservers = [
    "10.0.0.1"
  ]

  # The talos image to install
  talos_image = "ghcr.io/siderolabs/installer:v1.0.4"

  # Bootstrap Etcd as part of the creation process
  bootstrap = true

  # The base config from the node's talos_configuration
  # Contains shared information and secrets
  base_config = talos_configuration.single_example.base_config
}

# Create a local talosconfig
resource "local_file" "talosconfig" {
  content = talos_configuration.single_example.talosconfig
  filename = "talosconfig"
}

terraform {
  required_providers {
    talos = {
      source  = "localhost/j-lgs/talos"
      version = "0.0.10"
    }
    libvirt = {
      source = "dmacvicar/libvirt"
      version = "0.6.14"
    }
    macaddress = {
      source = "ivoronin/macaddress"
      version = "0.3.0"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_volume" "talos" {
  source = "https://github.com/siderolabs/talos/releases/download/v1.0.4/talos-amd64.iso"
  name   = "talos"
}

resource "libvirt_volume" "single_example_boot" {
  name   = "single_example_boot.qcow2"
  size   = 8589934592 # 4GiB
}

resource "libvirt_domain" "single_example" {
  name = "single_example"

  memory = "2048"
  vcpu   = 2

  disk {
    volume_id = libvirt_volume.talos.id
  }

  disk {
    volume_id = libvirt_volume.single_example_boot.id
  }

  network_interface {
    network_name   = "default"
    hostname       = "single_example"
    wait_for_lease = true
  }
}

resource "talos_configuration" "single_example" {
  # Talos configuration version target
  target_version = "v1.0"
  # Name of the talos cluster
  cluster_name = "taloscluster"
  # List of control plane nodes to act as endpoints the talos client should connect to
  endpoints = ["192.168.122.100"]

  # The evential endpoint that the kubernetes client will connect to
  kubernetes_endpoint = "https://192.168.122.100:6443"

  # The kubernetes version to be deployed
  kubernetes_version = "1.23.6"
}

resource "talos_control_node" "single_example" {
  # The node's hostname
  name = "cluster-control-1"

  # MAC address for the node. Will be used to apply the initial configuration
  macaddr = libvirt_domain.single_example.network_interface[0].mac
  # Expected network range to find the node's interface identified by `macaddr`
  dhcp_network_cidr = "192.168.122.0/25"
  # The disk to install talos onto
  install_disk = "/dev/vdb"

  # The primary interface's address in CIDR form
  ip = "192.168.122.100/24"
  # The primary interface's gateway address
  gateway = "192.168.122.1"
  # The node's nameservers
  nameservers = [
    "192.168.122.1"
  ]

  # Wireguard interface's eddress
  wg_address = "10.123.0.10/24"

  # IP addresses inside the network specified connect to the wireguard interface
  wg_allowed_ips = "10.123.0.0/24"

  # Wireguard endpoint for controlplane node to connect to
  wg_endpoint = "127.0.0.1:40000"

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

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


locals {
  nodes = {
    worker-node:  {ip: "192.168.122.201/24"}
    control-node: {ip: "192.168.122.200/24"}
  }
}

resource "libvirt_volume" "talos" {
  for_each = local.nodes
  source = "https://github.com/siderolabs/talos/releases/download/v1.0.4/talos-amd64.iso"
  name   = join("", ["talos_", each.key])
}

resource "macaddress" "worker_example" {
    for_each = local.nodes
}

resource "libvirt_volume" "worker_example_boot" {
  for_each = local.nodes
  name   = join("", [each.key, "_worker_example_boot.qcow2"])
  size   = 8589934592 # 8GiB
}

resource "libvirt_domain" "worker_example" {
  for_each = local.nodes
  name = each.key

  memory = "2048"
  vcpu   = 2

  disk {
    volume_id = libvirt_volume.talos[each.key].id
  }

  disk {
    volume_id = libvirt_volume.worker_example_boot[each.key].id
  }

  network_interface {
    network_name   = "default"
    hostname       = "worker_example"
    mac            = upper(macaddress.worker_example[each.key].address)
    wait_for_lease = true
  }
}

resource "talos_configuration" "worker_example" {
  # Talos configuration version target
  target_version = "v1.0"
  # Name of the talos cluster
  cluster_name = "taloscluster"
  # List of control plane nodes to act as endpoints the talos client should connect to
  endpoints = ["192.168.122.200"]

  # The evential endpoint that the kubernetes client will connect to
  kubernetes_endpoint = "https://192.168.122.200:6443"

  # The kubernetes version to be deployed
  kubernetes_version = "1.23.6"
}

resource "talos_control_node" "worker_example" {
  # The node's hostname
  name = "control-node"

  # MAC address for the node. Will be used to apply the initial configuration
  macaddr = libvirt_domain.worker_example["control-node"].network_interface[0].mac
  # Expected network range to find the node's interface identified by `macaddr`
  dhcp_network_cidr = "192.168.122.0/25"
  # The disk to install talos onto
  install_disk = "/dev/vdb"

  # The primary interface's address in CIDR form
  ip = "192.168.122.200/24"
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
  base_config = talos_configuration.worker_example.base_config
}

resource "talos_worker_node" "worker_example" {
  # The node's hostname
  name = "worker-node"

  # MAC address for the node. Will be used to apply the initial configuration
  macaddr = libvirt_domain.worker_example["worker-node"].network_interface[0].mac
  # Expected network range to find the node's interface identified by `macaddr`
  dhcp_network_cidr = "192.168.122.0/25"
  # The disk to install talos onto
  install_disk = "/dev/vdb"

  cluster_apiserver_args = {
    "allow-privileged" = "true"
  }

  kernel_args = ["i915.enable_guc=2", "i915.enable_dc=0"]

  interface {
    # The interface's name
    name = "eth0"
    # The interface's addresses in CIDR form
    addresses = [
      "192.168.122.201/24"
    ]
    route {
      network = "0.0.0.0/0"
      gateway = "192.168.122.1"
    }
  }

  cluster_proxy_args = {
    "ipvs-strict-arp" = "true"
  }

  sysctls = {
    "vm.nr_hugepages": "128"
  }

  kubelet_extra_mount {
    destination = "/var/local"
    type = "bind"
    source = "/var/local"
    options = [
      "rbind",
      "rshared",
      "rw"
    ]
  } 
  
  kubelet_extra_args = {
    "node-labels": "openebs.io/engine=mayastor"
  }
  
  udev = [
    "SUBSYSTEM==\"drm\", KERNEL==\"renderD*\", GROUP=\"103\", MODE=\"0666\"",
    "SUBSYSTEM==\"drm\", KERNEL==\"card*\",    GROUP=\"44\",  MODE=\"0666\""
  ]
  
  # The talos image to install
  talos_image = "ghcr.io/siderolabs/installer:v1.0.4"

  # The node's nameservers
  nameservers = [
    "192.168.122.1"
  ]

  base_config = talos_configuration.worker_example.base_config
}

# Create a local talosconfig
resource "local_file" "talosconfig" {
  content = talos_configuration.worker_example.talosconfig
  filename = "talosconfig"
}

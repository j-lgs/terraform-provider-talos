terraform {
  required_providers {
    talos = {
      source  = "localhost/j-lgs/talos"
      version = "0.0.5"
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
    # Node names MUST be lowercase RFC 1123 subdomains
    control-node-1: {ip = "192.168.122.100/24", wg_ip = "10.123.0.10/24" }
    control-node-2: {ip = "192.168.122.101/24", wg_ip = "10.123.0.11/24" }
    control-node-3: {ip = "192.168.122.102/24", wg_ip = "10.123.0.12/24" }
  }
  priorities = [
    {state: "MASTER", priority: 200},
    {state: "BACKUP", priority: 50},
    {state: "BACKUP", priority: 25}
  ]
  prioritymap = zipmap(keys(local.nodes), local.priorities)
  bootstrap_node_key = random_shuffle.nodes.result[0]
  bootstrap_node = local.nodes[random_shuffle.nodes.result[0]]
}

resource "libvirt_volume" "talos" {
  for_each = local.nodes
  source = "https://github.com/siderolabs/talos/releases/download/v1.0.4/talos-amd64.iso"
  name   = join("", ["talos_", each.key])
}

resource "macaddress" "ha_example" {
    for_each = local.nodes
}

resource "random_password" "vip_pass" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "random_shuffle" "nodes" {
  input        = keys(local.nodes)
}

resource "libvirt_volume" "ha_example_boot" {
  for_each = local.nodes
  name   = join("", [each.key, "_ha_example_boot.img"])
  size   = 8589934592 # 8GiB
}

resource "libvirt_domain" "ha_example" {
  for_each = local.nodes
  name = each.key

  memory = "2048"
  vcpu   = 2

  disk {
    volume_id = libvirt_volume.talos[each.key].id
  }

  disk {
    volume_id = libvirt_volume.ha_example_boot[each.key].id
  }

  network_interface {
    network_name   = "default"
    hostname       = each.key
    mac            = upper(macaddress.ha_example[each.key].address)
    wait_for_lease = true
  }
}

resource "talos_configuration" "ha_example" {
  # Talos configuration version target
  target_version = "v1.0"
  # Name of the talos cluster
  cluster_name = "taloscluster"
  # List of control plane nodes to act as endpoints the talos client should connect to
  endpoints = ["192.168.122.100", "192.168.122.101", "192.168.122.102"]

  # The evential endpoint that the kubernetes client will connect to
  kubernetes_endpoint = "https://192.168.122.150:6443"

  # The kubernetes version to be deployed
  kubernetes_version = "1.23.6"
}

resource "talos_control_node" "ha_example" {
  # The node's hostname
  for_each = local.nodes
  name = each.key

  # MAC address for the node. Will be used to apply the initial configuration
  macaddr = libvirt_domain.ha_example[each.key].network_interface[0].mac
  # Expected network range to find the node's interface identified by `macaddr`
  dhcp_network_cidr = "192.168.122.0/25"
  # The disk to install talos onto
  install_disk = "/dev/vdb"

  # The primary interface's address in CIDR form
  ip = each.value.ip
  # The primary interface's gateway address
  gateway = "192.168.122.1"
  # The node's nameservers
  nameservers = [
    "192.168.122.1"
  ]

  # Virtual IP address the shared API proxy will be served over
  api_proxy_ip = "192.168.122.150"

  # List of peers for the keepalived and haproxy cluster
  peers = [for _, n in setsubtract([for _, n in local.nodes : n.ip], [each.value.ip]) : split("/", n)[0]]

  # Bootstrap Etcd as part of the creation process
  bootstrap = local.bootstrap_node_key == each.key

  router_id = "11"
  state    = local.prioritymap[each.key].state
  priority = local.prioritymap[each.key].priority
  vip_pass = random_password.vip_pass.result

  # Wireguard interface's eddress
  wg_address = each.value.wg_ip

  # IP addresses inside the network specified connect to the wireguard interface
  wg_allowed_ips = "10.123.0.0/2"

  # Wireguard endpoint for controlplane node to connect to
  wg_endpoint = "127.0.0.1:40000"

  # The talos image to install
  talos_image = "ghcr.io/siderolabs/installer:v1.0.4"

  # The base config from the node's talos_configuration
  # Contains shared information and secrets
  base_config = talos_configuration.ha_example.base_config
}

# Create a local talosconfig
resource "local_file" "talosconfig" {
  content = talos_configuration.ha_example.talosconfig
  filename = "talosconfig"
}

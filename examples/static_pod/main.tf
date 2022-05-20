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
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_volume" "talos" {
  source = "https://github.com/siderolabs/talos/releases/download/v1.0.4/talos-amd64.iso"
  name   = "talos"
}

resource "libvirt_volume" "static_pod_boot" {
  name   = "static_pod_boot.qcow2"
  size   = 8589934592 # 4GiB
}

resource "libvirt_domain" "static_pod" {
  name = "static_pod"

  memory = "2048"
  vcpu   = 2

  disk {
    volume_id = libvirt_volume.talos.id
  }

  disk {
    volume_id = libvirt_volume.static_pod_boot.id
  }

  network_interface {
    network_name   = "default"
    hostname       = "static_pod"
    wait_for_lease = true
  }
}

resource "talos_configuration" "static_pod" {
  # Talos configuration version target
  target_version = "v1.0"
  # Name of the talos cluster
  name = "taloscluster"
  # List of control plane nodes to act as endpoints the talos client should connect to
  talos_endpoints = ["192.168.122.100"]

  # The evential endpoint that the kubernetes client will connect to
  kubernetes_endpoint = "https://192.168.122.100:6443"

  # The kubernetes version to be deployed
  kubernetes_version = "1.23.6"
}

resource "talos_control_node" "static_pod" {
  # The node's hostname
  name = "cluster-control-1"

  # MAC address for the node. Will be used to apply the initial configuration
  macaddr = libvirt_domain.static_pod.network_interface[0].mac
  # Expected network range to find the node's interface identified by `macaddr`
  dhcp_network_cidr = "192.168.122.128/25"
  # The disk to install talos onto
  install_disk = "/dev/vdb"

  devices = {
	# The interface's name
	"eth0": {
	  # The interface's addresses in CIDR form
	  addresses = [
		"192.168.122.100/24"
	  ]
	  routes = [{
		network = "0.0.0.0/0"
		gateway = "192.168.122.1"
	  }]
	}
  }

  # The node's nameservers

  nameservers = [
    "192.168.122.1"
  ]


  kubelet = {
	extra_mount = [{
      destination = "/var/static-confs"
      type = "bind"
      source = "/var/static-confs"
      options = [
        "rbind",
        "rshared",
        "rw"
      ]
	}]
  }

  files = [{
	content = <<EOT
global
  log         /dev/log local0
  log         /dev/log local1 notice
  daemon
defaults
  mode                    tcp
  log                     global
  option                  tcplog
  option                  tcp-check
  option                  dontlognull
  retries                 3
  timeout client          20s
  timeout server          20s
  timeout check           10s
  timeout queue           20s
  option                  redispatch
  timeout connect         5s
frontend http_stats
  bind 192.168.122.100:8080
  mode http
  stats uri /haproxy?stats
EOT
	permissions = 438
	path = "/var/static-confs/haproxy/haproxy.cfg"
	op = "create"
  }]

  pod = [<<EOT
apiVersion: v1
kind: Pod
metadata:
 name: haproxy
 namespace: kube-system
spec:
  containers:
  - image: haproxy:2.5.6
    name: haproxy-controlplane
    volumeMounts:
    - mountPath: /usr/local/etc/haproxy/haproxy.cfg
      name: haproxyconf
      readOnly: true
  hostNetwork: true
  volumes:
  - hostPath:
      path: /var/static-confs/haproxy/haproxy.cfg
      type: File
    name: haproxyconf
status: {}
EOT
  ]

  registry = {
	mirrors = {
	  "docker.io":  [ "http://172.17.0.1:5000" ],
	  "k8s.gcr.io": [ "http://172.17.0.1:5001" ],
	  "quay.io":    [ "http://172.17.0.1:5002" ],
	  "gcr.io":     [ "http://172.17.0.1:5003" ],
      "ghcr.io":    [ "http://172.17.0.1:5004" ],
    }
  }

  # The talos image to install
  talos_image = "ghcr.io/siderolabs/installer:v1.0.4"

  # Bootstrap Etcd as part of the creation process
  bootstrap = true
  bootstrap_ip = "192.168.122.100"

  # The base config from the node's talos_configuration
  # Contains shared information and secrets
  base_config = talos_configuration.static_pod.base_config
}

# Create a local talosconfig
resource "local_file" "talosconfig" {
  content = talos_configuration.static_pod.talos_config
  filename = "talosconfig"
}

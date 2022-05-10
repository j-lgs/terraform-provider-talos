package talos

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"text/template"
	"time"

	talosresource "github.com/talos-systems/talos/pkg/machinery/api/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

func testAccTalosConnectivity(path string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[path]
		if !ok {
			return fmt.Errorf("Not found: %s", path)
		}

		return testTalosConnectivity(rs, path, name)
	}
}

func genIPsNoCollision(ipbase string, min int, max int, n int) []string {
	ips := []string{}
	f := rand.Perm(max - min)

	for i := 0; i < n; i++ {
		ips = append(ips, ipbase+strconv.Itoa(min+f[i]))
	}

	return ips
}

func testAccKubernetesConnectivity(path string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[path]
		if !ok {
			return fmt.Errorf("Not found: %s", path)
		}

		is := rs.Primary
		cidr, ok := is.Attributes["interface.0.addresses.0"]
		if !ok {
			return fmt.Errorf("testTalosConnectivity: Unable to get interface 0, ip address 0, from resource")
		}
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("testTalosConnectivity: Must provide a valid CIDR IP address, got \"%s\", error \"%s\"", cidr, err.Error())
		}

		ctx := context.Background()
		// Wait for node to finish bootstrapping, ensure the kubernetes port is up before testing for it's health status
		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		httpclient := http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
		stats := "https://" + ip.String() + ":6443/readyz?verbose"

		var readyresp *http.Response
		for {
			r, err := httpclient.Get(stats)
			if err == nil && r.StatusCode >= 200 && r.StatusCode <= 299 {
				readyresp = r
				break
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf(ctx.Err().Error() + " - Reason - " + err.Error())
			default:
				time.Sleep(2 * time.Second)
			}
		}

		defer readyresp.Body.Close()

		return nil
	}
}

func testTalosConnectivity(rs *terraform.ResourceState, path string, name string) error {
	is := rs.Primary
	cidr, ok := is.Attributes["interface.0.addresses.0"]
	if !ok {
		return fmt.Errorf("testTalosConnectivity: Unable to get interface 0, ip address 0, from resource")
	}
	input_json, ok := is.Attributes["base_config"]
	if !ok {
		return fmt.Errorf("testTalosConnectivity: Unable to get base_config from resource")
	}

	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("testTalosConnectivity: Must provide a valid CIDR IP address, got \"%s\", error \"%s\"", cidr, err.Error())
	}

	input := generate.Input{}
	if err := json.Unmarshal([]byte(input_json), &input); err != nil {
		return err
	}

	host := net.JoinHostPort(ip.String(), strconv.Itoa(talos_port))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	conn, diags := secureConn(ctx, input, host)
	if diags != nil {
		return fmt.Errorf("testTalosConnectivity: Unable to connect to talos API at %s, maybe timed out.", host)
	}
	defer conn.Close()

	client := talosresource.NewResourceServiceClient(conn)
	resp, err := client.Get(ctx, &talosresource.GetRequest{
		Type:      "MachineConfig",
		Namespace: "config",
		Id:        "v1alpha1",
	})
	if err != nil {
		return fmt.Errorf("Error getting machine configuration, error \"%s\"", err.Error())
	}

	if len(resp.Messages) < 1 {
		return fmt.Errorf("Invalid message count recieved. Expected > 1 but got %d", len(resp.Messages))
	}

	return nil
}

var (
	gateway      string = "192.168.122.1"
	installdisk  string = "/dev/vdb"
	nameserver   string = "192.168.122.1"
	ip_worker_1  string = "192.168.122.110"
	ip_control_1 string = "192.168.122.121"
	ip_control_2 string = "192.168.122.122"
	ip_control_3 string = "192.168.122.123"
	dhcp_cidr    string = "192.168.122.128/25"
	talos_image  string = "ghcr.io/siderolabs/installer:v1.0.4"
)

func workerResource_basic(vmName string, rName string, ip string, additionals ...string) string {
	m := map[string]interface{}{
		"name":       rName,
		"vm_name":    vmName,
		"dhcp_cidr":  dhcp_cidr,
		"disk":       installdisk,
		"ip":         ip,
		"gateway":    gateway,
		"image":      talos_image,
		"nameserver": nameserver,
		"additional": "",
	}
	if len(additionals) > 0 {
		m["additional"] = additionals[0]
	}
	t := template.Must(template.New("").Parse(`
resource "libvirt_volume" "talos_{{ .vm_name }}_{{ .name }}" {
  source = "http://localhost:8000/talos-amd64.iso"
  name   = "test_talos-{{ .vm_name }}"
}

resource "libvirt_volume" "node_boot_{{ .vm_name }}_{{ .name }}" {
  name   = "test_{{ .vm_name }}_node_boot.qcow2"
  size   = 4294967296 # 4GiB
}

resource "libvirt_domain" "node_{{ .vm_name }}_{{ .name }}" {
  name = "test_{{ .vm_name }}_{{ .name }}"

  memory = "2048"
  vcpu   = 2

  disk {
    volume_id = libvirt_volume.talos_{{ .vm_name }}_{{ .name }}.id
  }

  disk {
    volume_id = libvirt_volume.node_boot_{{ .vm_name }}_{{ .name }}.id
  }

  network_interface {
    network_name   = "default"
    hostname       = "{{ .name }}"
    wait_for_lease = true
  }
}

resource "talos_worker_node" "{{ .name }}" {
  name = "{{ .name }}"
  macaddr = libvirt_domain.node_{{ .vm_name }}_{{ .name }}.network_interface[0].mac

  dhcp_network_cidr = "{{ .dhcp_cidr }}"
  install_disk = "{{ .disk }}"

  interface {
    name = "eth0"
    addresses = [
      "{{ .ip }}/24"
    ]
    route {
      network = "0.0.0.0/0"
      gateway = "{{ .gateway }}"
    }
  }

  talos_image = "{{ .image }}"
  nameservers = [
    "{{ .nameserver }}"
  ]

  registry_mirrors = {
    "docker.io": "http://172.17.0.1:5000",
    "k8s.gcr.io": "http://172.17.0.1:5001",
    "quay.io": "http://172.17.0.1:5002",
    "gcr.io": "http://172.17.0.1:5003",
    "ghcr.io": "http://172.17.0.1:5004"
  }

  {{ .additional }}

  base_config = talos_configuration.cluster.base_config
}

`))
	var b bytes.Buffer
	t.Execute(&b, m)
	return b.String()
}

//  libvirt_domain.node.network_interface[0].mac
func controlResource_basic(vmName string, rName string, bootstrap bool, ip string, additionals ...string) string {
	m := map[string]interface{}{
		"name":       rName,
		"vm_name":    vmName,
		"bootstrap":  strconv.FormatBool(bootstrap),
		"dhcp_cidr":  dhcp_cidr,
		"disk":       installdisk,
		"ip":         ip,
		"gateway":    gateway,
		"image":      talos_image,
		"nameserver": nameserver,
		"additional": "",
	}
	if len(additionals) > 0 {
		m["additional"] = additionals[0]
	}
	t := template.Must(template.New("").Parse(`
resource "libvirt_volume" "talos_{{ .vm_name }}_{{ .name }}" {
  source = "http://localhost:8000/talos-amd64.iso"
  name   = "test_talos-{{ .vm_name }}_{{ .name }}"
}

resource "libvirt_volume" "node_boot_{{ .vm_name }}_{{ .name }}" {
  name   = "test_{{ .vm_name }}_{{ .name }}_node_boot.qcow2"
  size   = 4294967296 # 4GiB
}

resource "libvirt_domain" "node_{{ .vm_name }}_{{ .name }}" {
  name = "test_{{ .vm_name }}_{{ .name }}"

  memory = "2048"
  vcpu   = 2

  disk {
    volume_id = libvirt_volume.talos_{{ .vm_name }}_{{ .name }}.id
  }

  disk {
    volume_id = libvirt_volume.node_boot_{{ .vm_name }}_{{ .name }}.id
  }

  network_interface {
    network_name   = "default"
    hostname       = "{{ .name }}"
    wait_for_lease = true
  }
}

resource "talos_control_node" "{{ .name }}" {
  name = "{{ .name }}"
  macaddr = libvirt_domain.node_{{ .vm_name }}_{{ .name }}.network_interface[0].mac

  bootstrap = {{ .bootstrap }}
  bootstrap_ip = "{{ .ip }}"

  dhcp_network_cidr = "{{ .dhcp_cidr }}"
  install_disk = "{{ .disk }}"

  interface {
    name = "eth0"
    addresses = [
      "{{ .ip }}/24"
    ]
    route {
      network = "0.0.0.0/0"
      gateway = "{{ .gateway }}"
    }
  }

  talos_image = "{{ .image }}"
  nameservers = [
    "{{ .nameserver }}"
  ]

  registry_mirrors = {
    "docker.io": "http://172.17.0.1:5000",
    "k8s.gcr.io": "http://172.17.0.1:5001",
    "quay.io": "http://172.17.0.1:5002",
    "gcr.io": "http://172.17.0.1:5003",
    "ghcr.io": "http://172.17.0.1:5004"
  }

  base_config = talos_configuration.cluster.base_config

  {{ .additional }}
}

`))
	var b bytes.Buffer
	t.Execute(&b, m)
	return b.String()
}

// testAccTalosWorker_basic creates the most basic talos worker configuration
func talosConfig_basic(endpoint string, kube_endpoint string, additionals ...string) string {
	m := map[string]interface{}{
		"endpoint":      endpoint,
		"kube_endpoint": kube_endpoint,
		"additional":    "",
	}
	if len(additionals) > 0 {
		m["additional"] = additionals[0]
	}
	t := template.Must(template.New("").Parse(`
terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
      version = "0.6.14"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

resource "talos_configuration" "cluster" {
  target_version = "v1.0"
  cluster_name = "taloscluster"
  endpoints = ["{{ .endpoint }}"]
  kubernetes_endpoint = "{{ .kube_endpoint }}"
  kubernetes_version = "1.23.6"

  {{ .additional }}
}

`))
	var b bytes.Buffer
	t.Execute(&b, m)
	return b.String()
}

func tes() string {
	return `
terraform {
  required_providers {
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
  source = "http://localhost:8000/talos-amd64.iso"
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
  dhcp_network_cidr = "192.168.122.128/25"
  # The disk to install talos onto
  install_disk = "/dev/vdb"

 interface {
    # The interface's name
    name = "eth0"
    # The interface's addresses in CIDR form
    addresses = [
      "192.168.122.100/24"
    ]
    route {
      network = "0.0.0.0/0"
      gateway = "192.168.122.1"
    }
  }

  # The node's nameservers
  nameservers = [
    "192.168.122.1"
  ]

  registry_mirrors = {
    "docker.io": "http://172.17.0.1:5000",
    "k8s.gcr.io": "http://172.17.0.1:5001",
    "quay.io": "http://172.17.0.1:5002",
    "gcr.io": "http://172.17.0.1:5003",
    "ghcr.io": "http://172.17.0.1:5004"
  }

  # The talos image to install
  talos_image = "ghcr.io/siderolabs/installer:v1.0.4"

  # Bootstrap Etcd as part of the creation process
  bootstrap = true
  bootstrap_ip = "192.168.122.100"

  # The base config from the node's talos_configuration
  # Contains shared information and secrets
  base_config = talos_configuration.single_example.base_config
}
`
}

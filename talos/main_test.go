package talos

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

// Helper global literals defining constants used in test VM instantiation.
var (
	testNworkers int    = 1
	testNcontrol int    = 3
	testDisk4GiB string = "4294967296"
	//testDisk8GiB string = "8589934592"
	testMem2GiB  string = "2048"
	testNodevcpu string = "2"
	testHostname string = "node"

	talosIsoVersion   string = "v1.0.5"
	talosIsoFile      string = ".talos-amd64-" + talosIsoVersion + ".iso"
	talosIsoURL       string = "https://github.com/siderolabs/talos/releases/download/" + talosIsoVersion + "/talos-amd64.iso"
	talosLocalIsoPort string = "8099"
	talosLocalIsoURL  string = "http://localhost:" + talosLocalIsoPort + "/" + talosIsoFile

	testGlobalTimeout time.Duration = 30 * time.Minute
)

// testMACAddresses is a global variable that contains the mac addresses for all virtual machines.
// This is required by the provider for initially provisioning the machines.
var testMACAddresses = make([]string, testNcontrol+testNworkers)

// accTestNodeArgs specifies arguments for the accTestNodes template.
type accTestNodeArgs struct {
	Workers       int
	Controls      int
	Bootsize      string
	Memory        string
	Vcpu          string
	HostnameBase  string
	MachineLogDir string
	LocalIsoURL   string
	LibvirtCustom string
	DockerCustom  string
	RegistryMount string
	InitialIPs    []string
}

// accTestNodes contains a Terraform configuration that manages the VMs used for
// acceptance testing the talos provider.
// TODO: Automatically expose container ports and use them in tests
var accTestNodes string = `
terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
      version = "0.6.14"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "2.16.0"
    }
    macaddress = {
      source = "ivoronin/macaddress"
      version = "0.3.0"
    }
  }
}

provider "docker" {
{{if ne .DockerCustom ""}}
{{.DockerCustom}}
{{else}}
  host = "unix:///var/run/docker.sock"
{{end}}
}

resource "docker_image" "registry_2" {
  name = "registry:2"
}

resource "docker_image" "registry_2_5" {
  name = "registry:2.5"
}

resource "docker_container" "docker_mirror" {
  image = docker_image.registry_2.latest
  name  = "docker_mirror"
  env = [
    "REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io",
    "REGISTRY_HTTP_ADDR=0.0.0.0:80",
{{if ne .RegistryMount ""}}
    "REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/registry/docker.io",
{{end}}
  ]
  ports {
    internal = 80
    external = 55000
  }
{{if ne .RegistryMount ""}}
  mounts {
    type = "bind"
    target = "/registry"
    source = "{{.RegistryMount}}"
  }
{{end}}
}

resource "docker_container" "k8s_gcr_mirror" {
  image = docker_image.registry_2.latest
  name  = "k8s_gcr_mirror"
  env = [
    "REGISTRY_PROXY_REMOTEURL=https://k8s.gcr.io",
    "REGISTRY_HTTP_ADDR=0.0.0.0:80",
{{if ne .RegistryMount ""}}
    "REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/registry/k8s.gcr.io",
{{end}}
  ]
  ports {
    internal = 80
    external = 55001
  }
{{if ne .RegistryMount ""}}
  mounts {
    type = "bind"
    target = "/registry"
    source = "{{.RegistryMount}}"
  }
{{end}}
}

resource "docker_container" "quay_mirror" {
  image = docker_image.registry_2_5.latest
  name  = "quay_mirror"
  env = [
    "REGISTRY_PROXY_REMOTEURL=https://quay.io",
    "REGISTRY_HTTP_ADDR=0.0.0.0:80",
{{if ne .RegistryMount ""}}
    "REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/registry/quay.io",
{{end}}
  ]
  ports {
    internal = 80
    external = 55002
  }
{{if ne .RegistryMount ""}}
  mounts {
    type = "bind"
    target = "/registry"
    source = "{{.RegistryMount}}"
  }
{{end}}
}

resource "docker_container" "gcr_mirror" {
  image = docker_image.registry_2.latest
  name  = "gcr_mirror"
  env = [
    "REGISTRY_PROXY_REMOTEURL=https://gcr.io",
    "REGISTRY_HTTP_ADDR=0.0.0.0:80",
{{if ne .RegistryMount ""}}
    "REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/registry/gcr.io",
{{end}}
  ]
  ports {
    internal = 80
    external = 55003
  }
{{if ne .RegistryMount ""}}
  mounts {
    type = "bind"
    target = "/registry"
    source = "{{.RegistryMount}}"
  }
{{end}}
}

resource "docker_container" "ghcr_mirror" {
  image = docker_image.registry_2.latest
  name  = "ghcr_mirror"
  env = [
    "REGISTRY_PROXY_REMOTEURL=https://ghcr.io",
    "REGISTRY_HTTP_ADDR=0.0.0.0:80",
{{if ne .RegistryMount ""}}
    "REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/registry/ghcr.io",
{{end}}
  ]
  ports {
    internal = 80
    external = 55004
  }
{{if ne .RegistryMount ""}}
  mounts {
    type = "bind"
    target = "/registry"
    source = "{{.RegistryMount}}"
  }
{{end}}
}

provider "libvirt" {
{{if ne .LibvirtCustom ""}}
{{.LibvirtCustom}}
{{else}}
  uri = "qemu:///system"
{{end}}
}

locals {
  n_workers = {{.Workers}}
  n_control = {{.Controls}}
  n_nodes   = local.n_workers + local.n_control
{{if ne .MachineLogDir ""}}
  file_console = [for i in range(local.n_nodes): <<TOC
<?xml version="1.0" ?>
<xsl:stylesheet version="1.0"
                xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output omit-xml-declaration="yes" indent="yes"/>
  <xsl:template match="node()|@*">
     <xsl:copy>
       <xsl:apply-templates select="node()|@*"/>
     </xsl:copy>
  </xsl:template>

  <xsl:template match="/domain/devices">
    <xsl:copy>
      <xsl:apply-templates select="node()|@*"/>
      <xsl:element name="serial">
        <xsl:attribute name="type">file</xsl:attribute>
        <xsl:element name="source">
          <xsl:attribute name="path">{{.MachineLogDir}}/.machine-${i}.log</xsl:attribute>
        </xsl:element>
        <xsl:element name="target">
          <xsl:attribute name="type">isa-serial</xsl:attribute>
          <xsl:attribute name="port">1</xsl:attribute>
          <xsl:element name="model">
            <xsl:attribute name="name">isa-serial</xsl:attribute>
          </xsl:element>
        </xsl:element>
      </xsl:element>
    </xsl:copy>
  </xsl:template>
</xsl:stylesheet>
TOC
  ]
{{end}}
  nethosts = <<TOC
<?xml version="1.0" encoding="UTF-8" ?>
<xsl:transform version="2.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output omit-xml-declaration="yes" indent="yes"/>
  <xsl:template match="node()|@*">
     <xsl:copy>
       <xsl:apply-templates select="node()|@*"/>
     </xsl:copy>
  </xsl:template>

  <xsl:template match="/network/ip/dhcp">
    <xsl:copy>
      <xsl:apply-templates select="node()|@*"/>
{{range $index, $ip := .InitialIPs}} 
      <xsl:element name="host">
        <xsl:attribute name="mac">${macaddress.leases[{{$index}}].address}</xsl:attribute>

        <xsl:attribute name="name">node-{{$index}}</xsl:attribute>

        <xsl:attribute name="ip">{{$ip}}</xsl:attribute>
      </xsl:element>
{{end}}
    </xsl:copy>
  </xsl:template>
</xsl:transform>
TOC
}

resource "macaddress" "leases" {
  count  = local.n_nodes
  prefix = [82,84,00]
}

resource "libvirt_network" "talos_network" {
  name = "talos-acctest"
  domain = "talos-acctest.local"
  addresses = ["192.168.124.0/24"]
  dns {
    enabled = true
  }
  xml {
    xslt = local.nethosts
  }
}

resource "libvirt_volume" "iso" {
  count  = local.n_nodes
  source = "{{.LocalIsoURL}}"
  name   = "test_talos_${count.index}"
}

resource "libvirt_volume" "boot" {
  count  = local.n_nodes
  name   = "test_boot_${count.index}.qcow2"
  size   = {{.Bootsize}}
}

resource "libvirt_domain" "test_node" {
  count = local.n_nodes
  name  = "test_control_${count.index}"

  memory = {{.Memory}}
  vcpu   = {{.Vcpu}}

  disk {
    volume_id = libvirt_volume.iso[count.index].id
  }

  disk {
    volume_id = libvirt_volume.boot[count.index].id
  }

  network_interface {
    network_name   = libvirt_network.talos_network.name
    hostname       = "{{.HostnameBase}}-${count.index}"
    wait_for_lease = true
    mac            = macaddress.leases[count.index].address
  }

  xml {
    xslt = local.file_console[count.index]
  }
}
`

// getiso downloads a talos iso if it doesn't already exist.
func getiso(ctx context.Context) error {
	// No checksum validation is performed.
	iso := "../" + talosIsoFile
	if _, err := os.Stat(iso); err != nil {
		iso, err := os.Create(iso)
		if err != nil {
			return err
		}
		defer iso.Close()

		resp, err := http.Get(talosIsoURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		_, err = io.Copy(iso, resp.Body)
		if err != nil {
			return err
		}

		return nil
	}

	log.Print("skipped grabbing the Talos iso since tests have been ran before.")

	return nil
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), testGlobalTimeout)
	defer cancel()
	var tf *tfexec.Terraform

	var doacc bool
	var err error

	if val, present := os.LookupEnv("TF_ACC"); present {
		doacc, err = strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("unable to parse boolean value for TF_ACC. got %s", val)
		}
	}
	if doacc {
		// Bringup of container registry and VMs
		log.Print("image cache containers and test VMs for acceptance testing will be created.")

		// Get Talos iso and serve it for VM template.
		getiso(ctx)

		isoServer := &http.Server{
			Handler: http.FileServer(http.Dir("../")),
			Addr:    ":" + talosLocalIsoPort,
		}
		go isoServer.ListenAndServe()
		log.Printf("Talos iso server is up on port %s\n", talosLocalIsoPort)

		installer := &releases.LatestVersion{
			Product: product.Terraform,
		}

		execPath, err := installer.Install(ctx)
		if err != nil {
			log.Fatalf("error installing Terraform: %s", err)
		}

		wd, err := ioutil.TempDir("", "talosProviderTest_")
		if err != nil {
			log.Fatalf("error creating temporary test directory: %s", err)
		}

		dir, exists := os.LookupEnv("MACHINELOG_DIR")
		if !exists {
			dir = "/tmp"
		}
		log.Print("note that machine logs are owned by root.\n")
		log.Printf("machine logs location: %s\n", dir)

		registry, exists := os.LookupEnv("REGISTRY_CACHE")
		if !exists {
			registry = ""
		}

		ips := []string{}
		for current := testInitialIPs.From(); current != testInitialIPs.To(); current = current.Next() {
			ips = append(ips, current.String())
		}

		// Template terraform configuration for test VMs.
		tpl := template.Must(template.New("").Parse(accTestNodes))
		var tfmain bytes.Buffer
		tpl.Execute(&tfmain, &accTestNodeArgs{
			Controls:      testNcontrol,
			Workers:       testNworkers,
			Bootsize:      testDisk4GiB,
			Memory:        testMem2GiB,
			Vcpu:          testNodevcpu,
			HostnameBase:  testHostname,
			LocalIsoURL:   talosLocalIsoURL,
			MachineLogDir: dir,
			RegistryMount: registry,
			InitialIPs:    ips,
		})

		tmpfn := filepath.Join(wd, "main.tf")
		if err := ioutil.WriteFile(tmpfn, tfmain.Bytes(), 0666); err != nil {
			log.Fatal(err)
		}

		defer os.RemoveAll(wd)

		// Run templated terraform configuration to create test VMs.
		tf, err = tfexec.NewTerraform(wd, execPath)
		if err != nil {
			log.Fatalf("error running NewTerraform: %s", err)
		}

		if err := tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
			log.Fatalf("error running Init: %s\nran using configuration:\n%s", err, tfmain.Bytes())
		}

		if err := tf.Apply(ctx); err != nil {
			log.Fatalf("error while creating test VMs: %s", err)
		}

		state, err := tf.Show(ctx)
		if err != nil {
			log.Fatalf("error running Show: %s", err)
		}

		// Get machine MAC addresses.
		nodes := []*tfjson.StateResource{}
		for _, res := range state.Values.RootModule.Resources {
			if res.Name == "test_node" {
				nodes = append(nodes, res)
			}
		}

		testMACAddresses = []string{}
		// First interface is used for setup. Should be kept in mind when creating test configurations.
		for _, node := range nodes {
			testMACAddresses = append(testMACAddresses,
				node.AttributeValues["network_interface"].([]interface{})[0].(map[string]interface{})["mac"].(string))
		}

	}

	code := m.Run()

	if doacc {
		log.Print("Image cache containers and test VMs are being destroyed.")
		if err := tf.Destroy(ctx); err != nil {
			log.Fatalf("error while tearing down test VMs: %s", err)
		}
	}

	os.Exit(code)
}

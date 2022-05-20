package talos

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	talosresource "github.com/talos-systems/talos/pkg/machinery/api/resource"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

// Talos configuration related global variables
var (
	kubernetesVersion string = "1.23.6"
)

// testConfig defines input variables needed when templating talos_configuration resource
// Terraform configurations.
type testConfig struct {
	Endpoint     string
	KubeEndpoint string
	KubeVersion  string
	TalosConfig  string
}

// talosConfig templates a Terraform configuration describing a talos_configuration resource.
func testTalosConfig(config *testConfig) string {
	if config.KubeEndpoint == "" {
		config.KubeEndpoint = "https://" + config.Endpoint + ":6443"
	}

	dir, exists := os.LookupEnv("TALOSCONF_DIR")
	if !exists {
		dir = "/tmp"
	}
	config.TalosConfig = filepath.Join(dir, ".talosconfig")
	log.Printf("talos: %s\n", filepath.Join(dir, "talosconfig"))

	config.KubeVersion = kubernetesVersion
	t := template.Must(template.New("").Parse(`

resource "talos_configuration" "cluster" {
  target_version = "v1.0"
  name = "taloscluster"
  talos_endpoints = ["{{.Endpoint}}"]
  kubernetes_endpoint = "{{.KubeEndpoint}}"
  kubernetes_version = "{{.KubeVersion}}"
}

resource "local_file" "talosconfig" {
  filename = "{{.TalosConfig}}"
  content = talos_configuration.cluster.talos_config
}

`))
	var b bytes.Buffer
	t.Execute(&b, config)
	return b.String()
}

// Control and worker node related global variables.
var (
	installDisk  string = "/dev/vdb"
	talosVersion string = "v1.0.5"
	installImage string = "ghcr.io/siderolabs/installer:" + talosVersion
	setupNetwork string = "192.168.124.0/24"
	gateway      string = "192.168.124.1"
	nameserver   string = "192.168.124.1"
)

// testNode defines input variables needed when templating talos_*_node resource
// Terraform configurations.
type testNode struct {
	// Required
	IP        string
	Index     int
	MAC       string
	Bootstrap bool

	// Optional
	Disk             string
	Image            string
	NodeSetupNetwork string
	Nameserver       string
	Gateway          string
}

// testControlConfig
func testControlConfig(nodes ...*testNode) string {
	tpl := `resource "talos_control_node" "control_{{.Index}}" {
  name = "control-{{.Index}}"
  macaddr = "{{.MAC}}"

  bootstrap = {{.Bootstrap}}
  bootstrap_ip = "{{.IP}}"

  dhcp_network_cidr = "{{.NodeSetupNetwork}}"
  install_disk = "{{.Disk}}"

  devices = {
    "eth0" : {
      addresses = [
        "{{.IP}}/24"
      ]
      routes = [{
        network = "0.0.0.0/0"
        gateway = "{{.Gateway}}"
      }]
    }
  }

  talos_image = "{{.Image}}"
  nameservers = [
    "{{.Nameserver}}"
  ]

  registry = {
	mirrors = {
      "docker.io":  [ "http://172.17.0.1:55000" ],
      "k8s.gcr.io": [ "http://172.17.0.1:55001" ],
      "quay.io":    [ "http://172.17.0.1:55002" ],
      "gcr.io":     [ "http://172.17.0.1:55003" ],
      "ghcr.io":    [ "http://172.17.0.1:55004" ],
    }
  }

  base_config = talos_configuration.cluster.base_config
}
`
	var config strings.Builder

	for _, n := range nodes {
		n.Disk = installDisk
		n.Image = installImage
		n.NodeSetupNetwork = setupNetwork
		n.Nameserver = nameserver
		n.Gateway = gateway
		t := template.Must(template.New("").Parse(tpl))

		t.Execute(&config, n)
	}

	return config.String()
}

func testControlNodePath(index int) string {
	return "talos_control_node.control_" + strconv.Itoa(index)
}

/*
func testWorkerNodePath(index int) string {
	return "talos_worker_node.worker_" + strconv.Itoa(index)
}
*/

var (
	talosConnectivityTimeout      time.Duration = 1 * time.Minute
	kubernetesConnectivityTimeout time.Duration = 3 * time.Minute
)

func testAccTalosConnectivity(nodeResourcePath string, talosIP string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, talosConnectivityTimeout)
		defer cancel()

		rs, ok := s.RootModule().Resources[nodeResourcePath]
		if !ok {
			return fmt.Errorf("not found: %s", nodeResourcePath)
		}
		is := rs.Primary

		input_json, ok := is.Attributes["base_config"]
		if !ok {
			return fmt.Errorf("testTalosConnectivity: Unable to get base_config from resource")
		}

		// Get Talos input bundle so we can connect to the Talos API endpoint and confirm a connection.
		input := generate.Input{}
		if err := json.Unmarshal([]byte(input_json), &input); err != nil {
			return err
		}

		host := net.JoinHostPort(talosIP, strconv.Itoa(talos_port))
		conn, diags := secureConn(ctx, input, host)
		if diags != nil {
			return fmt.Errorf("testTalosConnectivity: Unable to connect to talos API at %s, maybe timed out", host)
		}
		defer conn.Close()

		client := talosresource.NewResourceServiceClient(conn)
		resp, err := client.Get(ctx, &talosresource.GetRequest{
			Type:      "MachineConfig",
			Namespace: "config",
			Id:        "v1alpha1",
		})
		if err != nil {
			return fmt.Errorf("error getting machine configuration, error \"%s\"", err.Error())
		}

		if len(resp.Messages) < 1 {
			return fmt.Errorf("invalid message count recieved. Expected > 1 but got %d", len(resp.Messages))
		}

		return nil
	}
}

func testAccKubernetesConnectivity(kubernetesEndpoint string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, kubernetesConnectivityTimeout)
		defer cancel()

		httpclient := http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}

		var readyresp *http.Response
		for {
			r, err := httpclient.Get(kubernetesEndpoint + "/readyz?verbose")
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

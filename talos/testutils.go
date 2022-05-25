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

	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	clusterapi "github.com/talos-systems/talos/pkg/machinery/api/cluster"
	talosmachine "github.com/talos-systems/talos/pkg/machinery/api/machine"
	talosresource "github.com/talos-systems/talos/pkg/machinery/api/resource"
	"github.com/talos-systems/talos/pkg/machinery/client"
	clientconfig "github.com/talos-systems/talos/pkg/machinery/client/config"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"google.golang.org/grpc/codes"
)

// Talos configuration related global variables
var (
	testVersionContract          = config.TalosVersionCurrent
	testKubernetesVersion string = constants.DefaultKubernetesVersion
)

var configurationVersion string = "v1.0"

// testConfig defines input variables needed when templating talos_configuration resource
// Terraform configurations.
type testConfig struct {
	Endpoint     string
	ConfVersion  string
	KubeEndpoint string
	KubeVersion  string
	TalosConfig  string
}

var talosConfigName string = ".talosconfig"

// talosConfig templates a Terraform configuration describing a talos_configuration resource.
func testTalosConfig(config *testConfig) string {
	if config.KubeEndpoint == "" {
		config.KubeEndpoint = "https://" + config.Endpoint + ":6443"
	}

	dir, exists := os.LookupEnv("TALOSCONF_DIR")
	if !exists {
		dir = "/tmp"
	}
	config.TalosConfig = filepath.Join(dir, talosConfigName)
	log.Printf("talosconfig location: %s\n", filepath.Join(dir, talosConfigName))

	config.ConfVersion = configurationVersion

	config.KubeVersion = testKubernetesVersion
	t := template.Must(template.New("").Parse(`

resource "talos_configuration" "cluster" {
  target_version = "{{.ConfVersion}}"
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
	installDisk string = "/dev/vdb"
	//	talosVersion string = "v1.0.5"
	installImage string = generate.DefaultGenOptions().InstallImage
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

  install = {
    disk = "{{.Disk}}"
    image = "{{.Image}}"
  }

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

type testConnArg struct {
	resourcepath string
	talosIP      string
}

func ignoreErr(val any, err error) any {
	return val
}

func testAccTalosConnectivity(args ...testConnArg) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, arg := range args {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, talosConnectivityTimeout)
			defer cancel()

			rs, ok := s.RootModule().Resources[arg.resourcepath]
			if !ok {
				return fmt.Errorf("not found: %s", arg.talosIP)
			}
			is := rs.Primary

			inputJSON, ok := is.Attributes["base_config"]
			if !ok {
				return fmt.Errorf("testTalosConnectivity: Unable to get base_config from resource")
			}

			// Get Talos input bundle so we can connect to the Talos API endpoint and confirm a connection.
			input := generate.Input{}
			if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
				return err
			}

			host := net.JoinHostPort(arg.talosIP, strconv.Itoa(talosPort))
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

type clusterNodes struct {
	Control []string
	Worker  []string
}

func testAccTalosHealth(nodes *clusterNodes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		// total test duration context
		ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
		defer cancel()

		rs, ok := s.RootModule().Resources["talos_configuration.cluster"]
		if !ok {
			return fmt.Errorf("not found: %s", "talos_configuration.cluster")
		}
		is := rs.Primary

		config, ok := is.Attributes["talos_config"]
		if !ok {
			return fmt.Errorf("unable to get talos_config from resource")
		}

		cfg, err := clientconfig.FromString(config)
		if err != nil {
			return fmt.Errorf("error getting clientconfig form string: %w", err)
		}

		opts := []client.OptionFunc{
			client.WithConfig(cfg),
		}

		c, err := client.New(ctx, opts...)
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}
		defer c.Close()

		info := &clusterapi.ClusterInfo{}
		if len(nodes.Control) > 0 {
			info.ControlPlaneNodes = nodes.Control
		}
		if len(nodes.Worker) > 0 {
			info.WorkerNodes = nodes.Worker
		}

		for {
			err := testAccAttemptHealthcheck(ctx, info, c)
			if err == nil {
				return nil
			}
			select {
			case <-ctx.Done():
				return fmt.Errorf(ctx.Err().Error() + " - Reason - " + err.Error())
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func testAccAttemptHealthcheck(parentCtx context.Context, info *clusterapi.ClusterInfo, c *client.Client) error {
	// Wait for the node to fully setup.
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Minute)
	defer cancel()

	check, err := c.ClusterHealthCheck(ctx, 5*time.Minute, info)
	if err != nil {
		return fmt.Errorf("error getting healthcheck: %w", err)
	}

	if err := check.CloseSend(); err != nil {
		return fmt.Errorf("error running CloseSend on check: %w", err)
	}

	for {
		msg, err := check.Recv()
		if err != nil {
			if err == io.EOF || client.StatusCode(err) == codes.Canceled {
				return nil
			}
			return fmt.Errorf("error recieving checks: %w", err)
		}
		if msg.GetMetadata().GetError() != "" {
			return fmt.Errorf("healthcheck error: %s", msg.GetMetadata().GetError())
		}
	}
}

func testAccEnsureNMembers(membercount int, ip string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, talosConnectivityTimeout)
		defer cancel()

		rs, ok := s.RootModule().Resources["talos_configuration.cluster"]
		if !ok {
			return fmt.Errorf("not found: %s", "talos_configuration.cluster")
		}
		is := rs.Primary

		inputJSON, ok := is.Attributes["base_config"]
		if !ok {
			return fmt.Errorf("unable to get base_config from cluster resource")
		}

		// Get Talos input bundle so we can connect to the Talos API endpoint and confirm a connection.
		input := generate.Input{}
		if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
			return err
		}

		host := net.JoinHostPort(ip, strconv.Itoa(talosPort))
		conn, diags := secureConn(ctx, input, host)
		if diags != nil {
			return fmt.Errorf("testTalosConnectivity: Unable to connect to talos API at %s, maybe timed out", host)
		}
		defer conn.Close()

		client := talosmachine.NewMachineServiceClient(conn)
		resp, err := client.EtcdMemberList(ctx, &talosmachine.EtcdMemberListRequest{})
		if err != nil {
			return fmt.Errorf("error getting machine etcd members, error \"%s\"", err.Error())
		}

		if len(resp.Messages) < 1 {
			return fmt.Errorf("invalid message count recieved. expected > 1 but got %d", len(resp.Messages))
		}

		if count := len(resp.Messages[len(resp.Messages)-1].Members); count != membercount {
			return fmt.Errorf("invalid member count. expected %d but got %d", membercount, count)
		}

		return nil
	}
}

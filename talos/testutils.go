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
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/talos-systems/crypto/x509"
	clusterapi "github.com/talos-systems/talos/pkg/machinery/api/cluster"
	talosmachine "github.com/talos-systems/talos/pkg/machinery/api/machine"
	talosresource "github.com/talos-systems/talos/pkg/machinery/api/resource"
	"github.com/talos-systems/talos/pkg/machinery/client"
	clientconfig "github.com/talos-systems/talos/pkg/machinery/client/config"
	"github.com/talos-systems/talos/pkg/machinery/config"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/constants"
	"google.golang.org/grpc/codes"
)

// Talos configuration related global variables
var (
	testDefaultCerts = &generate.Certs{
		Admin: &x509.PEMEncodedCertificateAndKey{
			Crt: []byte(`-----BEGIN CERTIFICATE-----
MIIBPjCB8aADAgECAhBsTZO4+ItCrBhWmZtLnGzdMAUGAytlcDAQMQ4wDAYDVQQK
EwV0YWxvczAeFw0wOTExMTAyMzAwMDBaFw0xOTExMDgyMzAwMDBaMBAxDjAMBgNV
BAoTBXRhbG9zMCowBQYDK2VwAyEA3aj2K++3ZUm2jfZTk18ZT3i4Yun1XGnL443J
308kpBKjYTBfMA4GA1UdDwEB/wQEAwIChDAdBgNVHSUEFjAUBggrBgEFBQcDAQYI
KwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUEJycS3VNw11fpCdE
Tqw3Hiv9vqUwBQYDK2VwA0EAb4nWUXAGLvJXqRdBIZ9dAFjKNm+mOWQtfvFZPed+
ApSgVwZ08YhpTaldesUIDgsThcwMXV0GszRQV5Ponkn1Aw==
-----END CERTIFICATE-----
`),
			Key: []byte(`-----BEGIN ED25519 PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJCuZGq2GPWZnvtJvmwC+HIu6e95GecdBxC9qR4nGw4t
-----END ED25519 PRIVATE KEY-----
`),
		},
		Etcd: &x509.PEMEncodedCertificateAndKey{
			Crt: []byte(`-----BEGIN CERTIFICATE-----
MIIBfTCCASSgAwIBAgIRAMQ2ZBL8kPhDPGVnz6zLFXQwCgYIKoZIzj0EAwIwDzEN
MAsGA1UEChMEZXRjZDAeFw0wOTExMTAyMzAwMDBaFw0xOTExMDgyMzAwMDBaMA8x
DTALBgNVBAoTBGV0Y2QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQUOLEz7dLp
07Cpt0gSnhcXkp6OKhBvMBTIK4hsXLsWBUD8mVTgj1FfN/O/rByv9q2b5lpK3V4s
IMqaqyNwewWIo2EwXzAOBgNVHQ8BAf8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUH
AwEGCCsGAQUFBwMCMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFGPKZJSFBkFR
VoUViHMMHZ6r4Kb/MAoGCCqGSM49BAMCA0cAMEQCIDa/AGQdHywuRq0FKerzYu2H
mTgqTWOinhq09TfNyPBGAiBmjX4WoXLvrUP+rTdzgBXIscIIpiN9uCbYsTr3JFI5
7Q==
-----END CERTIFICATE-----
`),
			Key: []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAZuxEPxmC8Hx+bs3beez+Qes/KZtVny4iB2m7rbyX1YoAoGCCqGSM49
AwEHoUQDQgAEFDixM+3S6dOwqbdIEp4XF5KejioQbzAUyCuIbFy7FgVA/JlU4I9R
Xzfzv6wcr/atm+ZaSt1eLCDKmqsjcHsFiA==
-----END EC PRIVATE KEY-----
`),
		},
		K8s: &x509.PEMEncodedCertificateAndKey{
			Crt: []byte(`-----BEGIN CERTIFICATE-----
MIIBiDCCAS+gAwIBAgIQCt0vZrvsZ7x2bblRW1yYFTAKBggqhkjOPQQDAjAVMRMw
EQYDVQQKEwprdWJlcm5ldGVzMB4XDTA5MTExMDIzMDAwMFoXDTE5MTEwODIzMDAw
MFowFTETMBEGA1UEChMKa3ViZXJuZXRlczBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABDlXDE9L0/isznkWEFglmDgy6ygkDqSamT0Y65q8fca1f3FCZhibpVRMsAhu
73wr2y0Ovr4bIwcbC6mzsPelpRCjYTBfMA4GA1UdDwEB/wQEAwIChDAdBgNVHSUE
FjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4E
FgQU6xmnFNkBnNrrEB093YLuUFYbL/AwCgYIKoZIzj0EAwIDRwAwRAIgE0KSI2Tn
Z+uK7E+WTaXX2APYYYd9rr89jliYaQ4QE7gCICpuLOJty7zL0vt/yshPKWf+I38V
JhMskjqlPc9lEEG9
-----END CERTIFICATE-----
`),
			Key: []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIOS7Y3xPfreRGyDIXSO8huzoEcG7ZwlkpW2cM53WDULMoAoGCCqGSM49
AwEHoUQDQgAEOVcMT0vT+KzOeRYQWCWYODLrKCQOpJqZPRjrmrx9xrV/cUJmGJul
VEywCG7vfCvbLQ6+vhsjBxsLqbOw96WlEA==
-----END EC PRIVATE KEY-----
`),
		},
		K8sAggregator: &x509.PEMEncodedCertificateAndKey{
			Crt: []byte(`-----BEGIN CERTIFICATE-----
MIIBYDCCAQagAwIBAgIRAP5fAY5H6+fpF9wRlGE8u4YwCgYIKoZIzj0EAwIwADAe
Fw0wOTExMTAyMzAwMDBaFw0xOTExMDgyMzAwMDBaMAAwWTATBgcqhkjOPQIBBggq
hkjOPQMBBwNCAASaE+T1bLo06J1ZS9oGXXvEu4fI579YE2qyW+oH2O1ujpRVBx1r
DurTGt1aCRINfv9M3mxc2WI0elW2gGwiZDEto2EwXzAOBgNVHQ8BAf8EBAMCAoQw
HQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMA8GA1UdEwEB/wQFMAMBAf8w
HQYDVR0OBBYEFIVMh8KPOQXg763wjfDLqFyrrCUFMAoGCCqGSM49BAMCA0gAMEUC
IQDzDGJVNBx5YZWNJbd14OCYV6ghTyJAFy/7armaETE9pgIgHAdOQtMWVVyB/1Us
CIebtL7CcetnDVeaB4ijzG67lBc=
-----END CERTIFICATE-----
`),
			Key: []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIKi7D46Zo1XuuyrYiyMyiTAODdJeTPCGIUKPe6j9YTzYoAoGCCqGSM49
AwEHoUQDQgAEmhPk9Wy6NOidWUvaBl17xLuHyOe/WBNqslvqB9jtbo6UVQcdaw7q
0xrdWgkSDX7/TN5sXNliNHpVtoBsImQxLQ==
-----END EC PRIVATE KEY-----
`),
		},
		K8sServiceAccount: &x509.PEMEncodedKey{
			Key: []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIE2bYcsx3UqRIVr8F776i3X44PFfjNq3w5s4OxRvgA+doAoGCCqGSM49
AwEHoUQDQgAEjKovjKS75ObNPsyb2Ury9aP/dXZ9QHwereeXInWAlxzd3ctgDHQN
kGQ1kf6AOXlcAyBOb+KK0LnIh06QCUiZVg==
-----END EC PRIVATE KEY-----
`),
		},
		OS: &x509.PEMEncodedCertificateAndKey{
			Crt: []byte(`-----BEGIN CERTIFICATE-----
MIIBPjCB8aADAgECAhBsTZO4+ItCrBhWmZtLnGzdMAUGAytlcDAQMQ4wDAYDVQQK
EwV0YWxvczAeFw0wOTExMTAyMzAwMDBaFw0xOTExMDgyMzAwMDBaMBAxDjAMBgNV
BAoTBXRhbG9zMCowBQYDK2VwAyEA3aj2K++3ZUm2jfZTk18ZT3i4Yun1XGnL443J
308kpBKjYTBfMA4GA1UdDwEB/wQEAwIChDAdBgNVHSUEFjAUBggrBgEFBQcDAQYI
KwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUEJycS3VNw11fpCdE
Tqw3Hiv9vqUwBQYDK2VwA0EAb4nWUXAGLvJXqRdBIZ9dAFjKNm+mOWQtfvFZPed+
ApSgVwZ08YhpTaldesUIDgsThcwMXV0GszRQV5Ponkn1Aw==
-----END CERTIFICATE-----
`),
			Key: []byte(`-----BEGIN ED25519 PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJCuZGq2GPWZnvtJvmwC+HIu6e95GecdBxC9qR4nGw4t
-----END ED25519 PRIVATE KEY-----
`),
		},
	}
	testVersionContract      = config.TalosVersionCurrent
	testControlPlaneEndpoint = &v1alpha1.Endpoint{
		URL: &url.URL{
			Scheme:      "https",
			Opaque:      "",
			User:        nil,
			Host:        "test",
			Path:        "",
			RawPath:     "",
			ForceQuery:  false,
			RawQuery:    "",
			Fragment:    "",
			RawFragment: "",
		},
	}
	testCertSANs = []string{
		"0.0.0.0",
	}
	testSANs = []string{
		"hostname",
	}
	testClusterName   = "test"
	testClusterID     = "tSuqMd_jk2CU_wGDuPpE7A3HlY9_mcoXWWJ0kRbK8aE="
	testClusterSecret = "6SxVdcxHbbUdSsPpgnnRSHClbxkwmVpxNnbIKVGVirk="
	testServiceDomain = "cluster.local"
	testPodNet        = []string{
		"10.244.0.0/16",
	}
	testServiceNet = []string{
		"10.96.0.0/12",
	}
	testKubernetesVersion string = "1.23.6"

	testSecrets = &generate.Secrets{
		BootstrapToken:         "3pnxiu.jlzxqodjfujhyado",
		AESCBCEncryptionSecret: "64jlGeK1z13pY0NXKAo7VQHdzJRaugTdTZMflIZErTU=",
	}
	testTrustdInfo = &generate.TrustdInfo{
		Token: "s1pygp.a474wnneqo4v3lbs",
	}

	testExternalEtcd = false

	testInstallDisk  = "/dev/sda"
	testInstallImage = "test.io/talosinstall:latest"

	testInstallExtraKernelArgs = []string{
		"test",
	}
	testNetworkConfigOptions []v1alpha1.NetworkConfigOption = []v1alpha1.NetworkConfigOption{
		v1alpha1.WithNetworkInterfaceCIDR("eth0", "192.168.5/24"),
	}
	testCNIConfig = &v1alpha1.CNIConfig{
		CNIName: constants.FlannelCNI,
	}

	testRegistryMirror = map[string]*v1alpha1.RegistryMirrorConfig{
		"test.io": {
			MirrorEndpoints: []string{
				"test.org",
			},
		},
	}
	testRegistryConfig = map[string]*v1alpha1.RegistryConfig{
		"test.org": {
			RegistryTLS: &v1alpha1.RegistryTLSConfig{
				TLSClientIdentity: &x509.PEMEncodedCertificateAndKey{
					Crt: []byte("test"),
					Key: []byte("test"),
				},
				TLSCA:                 []byte("test"),
				TLSInsecureSkipVerify: false,
			},
			RegistryAuth: &v1alpha1.RegistryAuthConfig{
				RegistryUsername:      "username",
				RegistryPassword:      "password",
				RegistryAuth:          "auth",
				RegistryIdentityToken: "token",
			},
		},
	}
	testMachineDisks = []*v1alpha1.MachineDisk{
		{
			DeviceName: "/dev/sdb1",
			DiskPartitions: []*v1alpha1.DiskPartition{
				{
					DiskSize:       v1alpha1.DiskSize(100000000),
					DiskMountPoint: "/mnt",
				},
			},
		},
	}
	testSystemDiskEncryptionConfig = &v1alpha1.SystemDiskEncryptionConfig{
		StatePartition: &v1alpha1.EncryptionConfig{
			EncryptionProvider: "luks2",
			EncryptionKeys: []*v1alpha1.EncryptionKey{
				{
					KeyStatic: &v1alpha1.EncryptionKeyStatic{
						KeyData: "password",
					},
					KeySlot: 0,
				},
			},
			EncryptionCipher:    "aex-xts-plain64",
			EncryptionKeySize:   4096,
			EncryptionBlockSize: uint64(4096),
			EncryptionPerfOptions: []string{
				"same_cpu_crypt",
			},
		},
		EphemeralPartition: &v1alpha1.EncryptionConfig{
			EncryptionProvider: "luks2",
			EncryptionKeys: []*v1alpha1.EncryptionKey{
				{
					KeyNodeID: &v1alpha1.EncryptionKeyNodeID{},
					KeySlot:   0,
				},
			},
			EncryptionCipher:    "aex-xts-plain64",
			EncryptionKeySize:   4096,
			EncryptionBlockSize: uint64(4096),
			EncryptionPerfOptions: []string{
				"same_cpu_crypt",
			},
		},
	}
	testSysctls = map[string]string{
		"key": "value",
	}
	testDebug                    = false
	testPersist                  = true
	testAllowSchedulingOnMasters = true
	testDiscoveryEnabled         = true
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

type testConnArg struct {
	resourcepath string
	talosIP      string
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

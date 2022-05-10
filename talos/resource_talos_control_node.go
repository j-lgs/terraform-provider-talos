package talos

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/bits"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"crypto/tls"
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/instrumenta/kubeval/kubeval"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/api/resource"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	machinetype "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Controlplane specific schema
var (
	// ControlPlaneSchema contains machine specific configuration options.
	// See https://www.talos.dev/v1.0/reference/configuration/#machinecontrolplaneconfig for more information.
	ControlPlaneSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Machine specific configuration options.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"controller_manager_disabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
					Description: "Disable kube-controller-manager on the node.	",
				},
				"scheduler_disabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Disable kube-scheduler on the node.",
				},
			},
		},
	}

	AdmissionPluginConfigSchema schema.Resource = schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name is the name of the admission controller. It must match the registered admission plugin name.",
				// TODO Validate it is a properly formed name
			},
			"configuration": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Configuration is an embedded configuration object to be used as the plugin’s configuration.",
				// TODO Validate it is a properly formed YAML
			},
		},
	}

	// See https://www.talos.dev/v1.0/reference/configuration/#proxyconfig
	ProxyConfigSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Represents the kube proxy configuration options.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"image": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The container image used in the kube-proxy manifest.",
					ValidateFunc: validateImage,
				},
				"mode": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The container image used in the kube-proxy manifest.",
					Default:     "iptables",
					// TODO Validate it's a valid mode
				},
				"disabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Disable kube-proxy deployment on cluster bootstrap.",
					Default:     false,
				},
				"extra_args": Optional(StringMap("Extra arguments to supply to kube-proxy.")),
			},
		},
	}

	// See https://www.talos.dev/v1.0/reference/configuration/#controlplaneconfig
	ControlPlaneConfigSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Represents the control plane configuration options.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"endpoint": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.",
					// TODO Verify well formed endpoint
				},
				"local_api_server_port": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The port that the API server listens on internally. This may be different than the port portion listed in the endpoint field.",
					Default:     6443,
					// TODO Verify in correct port range
				},
			},
		},
	}

	// See https://www.talos.dev/v1.0/reference/configuration/#apiserverconfig
	ApiServerConfigSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Represents the kube apiserver configuration options.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"image": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The container image used in the API server manifest.",
					ValidateFunc: validateImage,
				},
				"extra_args": Optional(StringMap("Extra arguments to supply to the API server.")),

				"extra_volumes": {
					Type:        schema.TypeList,
					Optional:    true,
					MinItems:    1,
					Description: "Extra volumes to mount to the API server static pod.",
					Elem:        &VolumeMountSchema,
				},
				"env":       Optional(StringList("The env field allows for the addition of environment variables for the control plane component.")),
				"cert_sans": ValidateInner(Optional(StringList("Extra certificate subject alternative names for the API server’s certificate.")), validateIP),
				"disable_pod_security_policy": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Disable PodSecurityPolicy in the API server and default manifests.",
				},

				"admission_control": {
					Type:        schema.TypeList,
					MinItems:    1,
					Optional:    true,
					Description: "Configure the API server admission plugins.",
					Elem:        &AdmissionPluginConfigSchema,
				},
			},
		},
	}
)

func resourceControlNode() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceControlNodeCreate,
		ReadContext:   resourceControlNodeRead,
		UpdateContext: resourceControlNodeUpdate,
		DeleteContext: resourceControlNodeDelete,
		Schema: map[string]*schema.Schema{
			// Mandatory for minimal template generation
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateDomain,
				ForceNew:     true,
			},

			// Install arguments
			"install_disk": {
				Type:     schema.TypeString,
				Required: true,
			},
			"talos_image": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateImage,
			},
			"kernel_args": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"macaddr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateMAC,
			},
			"dhcp_network_cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDR,
			},

			// cluster arguments
			"cluster_apiserver_args": {
				Type:       schema.TypeMap,
				Deprecated: "Redundant",
				Optional:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cluster_proxy_args": {
				Type:       schema.TypeMap,
				Deprecated: "Redundant",
				Optional:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"local_apiserver_port": {
				Type:       schema.TypeString,
				Deprecated: "Redundant",
				Optional:   true,
				Default:    "",
			},

			// DEPRECATED registry_mirrors
			"registry_mirrors": {
				Deprecated: "Redundant",
				Type:       schema.TypeMap,
				Optional:   true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateEndpoint,
				},
			},

			// DEPRECATED Kubelet args
			"kubelet_extra_mount": &kubeletExtraMountSchema,
			"kubelet_extra_args": {
				Deprecated: "Redundant",
				Type:       schema.TypeMap,
				Optional:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// --- MachineConfig.
			// See https://www.talos.dev/v1.0/reference/configuration/#machineconfig for full spec.
			"cert_sans":     ValidateInner(Optional(StringList("Extra certificate subject alternative names for the machine’s certificate.")), validateIP),
			"control_plane": &ControlPlaneSchema,
			"kubelet":       &KubeletConfigSchema,
			"pod": {
				Type:        schema.TypeList,
				MinItems:    0,
				Optional:    true,
				Description: "Used to provide static pod definitions to be run by the kubelet directly bypassing the kube-apiserver.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
						v := value.(string)
						config := kubeval.NewDefaultConfig()
						schemaCache := kubeval.NewSchemaCache()
						_, err := kubeval.ValidateWithCache([]byte(v), schemaCache, config)
						if err != nil {
							errs = append(errs, fmt.Errorf("Invalid kubernetes manifest provided"))
							errs = append(errs, err)
						}
						return
					},
				},
			},
			// hostname derived from name
			"interface":   &networkInterfaceSchema,
			"nameservers": ValidateInner(Optional(StringList("Used to statically set the nameservers for the machine.")), validateEndpoint),
			"extra_host": {
				Type:        schema.TypeList,
				MinItems:    1,
				Optional:    true,
				Description: "Allows the addition of user specified files.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The IP of the host.",
						},
						"aliases": ValidateInner(Required(StringList("The host alias.")), validateHost),
					},
				},
			},
			// kubespan not implemented
			// disks not implemented
			// install not implemented
			"file": {
				Type:        schema.TypeList,
				MinItems:    1,
				Optional:    true,
				Description: "Allows the addition of user specified files.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The file's content. Not required to be base64 encoded.",
						},
						"permissions": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Unix permission for the file",
							ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
								v := value.(int)
								if v < 0 {
									errs = append(errs, fmt.Errorf("Persistent keepalive interval must be a positive integer, got %d", v))
								}
								return
							},
						},
						"path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Full path for the file to be created at.",
							// TODO: Add validation for path correctness
						},
						"op": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Mode for the file. Can be one of create, append and overwrite.",
							ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
								v := value.(string)
								switch v {
								case
									"create",
									"append",
									"overwrite":
									return
								default:
									errs = append(errs, fmt.Errorf("Invalid file op, must be one of \"create\", \"append\" or \"overwrite\", got %s", v))
								}
								return
							},
						},
					},
				},
			},
			"env": Optional(
				StringMap(
					"Allows for the addition of environment variables. All environment variables are set on PID 1 in addition to every service.")),
			// time not implemented
			"sysctls":  Optional(StringMap("Used to configure the machine’s sysctls.")),
			"sysfs":    Optional(StringMap("Used to configure the machine’s sysctls.")),
			"registry": &RegistryListSchema,
			// system_disk_encryption not implemented
			// features not implemented
			"udev": Optional(StringList("Configures the udev system.")),
			// logging not implemented
			// kernel not implemented
			// ----- MachineConfig End
			// ----- ClusterConfig Start
			"control_plane_config": &ControlPlaneConfigSchema,
			// clustername already filled
			// cluster_network not implemented
			"apiserver": &ApiServerConfigSchema,
			// controller manager not implemented
			"proxy": &ProxyConfigSchema,
			// scheduler not implemented
			// discovery not implemented
			// etcd not implemented
			// coredns not implemented
			// external_cloud_provider not implemented
			"extra_manifests": Optional(StringList("A list of urls that point to additional manifests. These will get automatically deployed as part of the bootstrap.")),
			// TODO Add verification function confirming it's a correct manifest that can be downloaded.
			// inline_manifests not implemented
			// admin_kubeconfig not implemented
			"allow_scheduling_on_masters": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Allows running workload on master nodes.",
			},
			// ----- ClusterConfig End

			// ----- Resource Cluster bootstrap configuration
			"bootstrap": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"bootstrap_ip": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateIP,
			},

			// From the cluster provider
			"base_config": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
					v := value.(string)
					input := generate.Input{}
					if err := json.Unmarshal([]byte(v), &input); err != nil {
						errs = append(errs, fmt.Errorf("Failed to parse base_config. Do not set this value to anything other than the base_config value of a talos_cluster_config resource"))
					}
					return
				},
			},

			// Generated
			"patch": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func checkArp(mac string) (net.IP, diag.Diagnostics) {
	arp, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, diag.Errorf("%s\n", err)
	}
	defer arp.Close()

	scanner := bufio.NewScanner(arp)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if strings.EqualFold(f[3], mac) {
			return net.ParseIP(f[0]), nil
		}
	}

	return nil, nil
}

func lookupIP(ctx context.Context, network string, mac string) (net.IP, diag.Diagnostics) {
	// Check if it's in the initial table

	ip := net.IP{}
	diags := diag.Diagnostics{}

	if ip, diags = checkArp(mac); diags != nil {
		return nil, diags
	}
	if ip != nil {
		return ip, diags
	}

	ctx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	for poll := true; poll; poll = (ip == nil) {
		select {
		case <-ctx.Done():
			return nil, diag.FromErr(ctx.Err())
		default:
			err := exec.CommandContext(ctx, "nmap", "-sP", network).Run()
			if err != nil {
				return nil, diag.Errorf("%s\n", err)
			}
			if ip, diags = checkArp(mac); diags != nil {
				return nil, diags
			}
			tflog.Error(ctx, ip.String())
			tflog.Error(ctx, mac)
			if ip != nil {
				return ip, nil
			}
			time.Sleep(5 * time.Second)
		}
	}

	return ip, nil
}

func makeTlsConfig(certs generate.Certs, secure bool) (tls.Config, diag.Diagnostics) {
	tlsConfig := &tls.Config{}
	if secure {
		clientCert, err := tls.X509KeyPair(certs.Admin.Crt, certs.Admin.Key)
		if err != nil {
			return tls.Config{}, diag.FromErr(err)
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(certs.OS.Crt); !ok {
			return tls.Config{}, diag.Errorf("Unable to append certs from PEM")
		}

		return tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{clientCert},
		}, nil
	}

	tlsConfig.InsecureSkipVerify = true
	return tls.Config{
		InsecureSkipVerify: true,
	}, nil

}

func waitTillTalosMachineUp(ctx context.Context, tlsConfig tls.Config, host string, secure bool) diag.Diagnostics {
	tflog.Info(ctx, "Waiting for talos machine to be up")
	// overall timeout should be 5 mins

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)),
		grpc.WithBlock(),
	}
	ctx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	for _, err := grpc.Dial(host, opts...); err != nil; {
		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		default:
			tflog.Info(ctx, "Retrying connection to "+host+" reason "+err.Error())
			time.Sleep(5 * time.Second)
		}
	}

	tflog.Info(ctx, "Waiting for talos machine to be up")

	return nil
}

func generateConfig(ctx context.Context, d *schema.ResourceData) ([]byte, diag.Diagnostics) {
	input := generate.Input{}
	if err := json.Unmarshal([]byte(d.Get("base_config").(string)), &input); err != nil {
		tflog.Error(ctx, "Failed to unmarshal input bundle: "+err.Error())
		return nil, diag.FromErr(err)
	}

	var controlCfg *v1alpha1.Config
	controlCfg, err := generate.Config(machinetype.TypeControlPlane, &input)
	if err != nil {
		tflog.Error(ctx, "failed to generate config for node: "+err.Error())
		return nil, diag.FromErr(err)
	}

	mc := controlCfg.MachineConfig
	cc := controlCfg.ClusterConfig

	mc.MachinePods = []v1alpha1.Unstructured{}
	for _, v := range d.Get("pod").([]interface{}) {
		var pod v1alpha1.Unstructured

		if err = yaml.Unmarshal([]byte(v.(string)), &pod); err != nil {
			tflog.Error(ctx, "failed to unmarshal static pod config into unstructured")
			return nil, diag.FromErr(err)
		}

		mc.MachinePods = append(mc.MachinePods, v1alpha1.Unstructured{
			Object: pod.Object,
		})
	}

	mc.MachineFiles = []*v1alpha1.MachineFile{}
	for _, v := range d.Get("file").([]interface{}) {
		file := v.(map[string]interface{})
		mc.MachineFiles = append(mc.MachineFiles, &v1alpha1.MachineFile{
			FileContent:     file["content"].(string),
			FilePermissions: v1alpha1.FileMode(file["permissions"].(int)),
			FilePath:        file["path"].(string),
			FileOp:          file["op"].(string),
		})
	}

	mc.MachineCertSANs = []string{}
	for _, v := range d.Get("cert_sans").([]interface{}) {
		mc.MachineCertSANs = append(mc.MachineCertSANs, v.(string))
	}

	port := d.Get("local_apiserver_port").(string)
	if port != "" {
		p := 0
		if p, err = strconv.Atoi(port); err != nil {
			return nil, diag.FromErr(err)
		}

		cc.ControlPlane.LocalAPIServerPort = p

	}

	if diags := generateCommonConfig(d, controlCfg); diags != nil {
		return []byte{}, diags
	}

	var controlYaml []byte

	controlYaml, err = controlCfg.Bytes()
	if err != nil {
		log.Fatalf("failed to generate config" + err.Error())
		return nil, diag.FromErr(err)
	}
	re := regexp.MustCompile(`\s*#.*`)
	no_comments := re.ReplaceAll(controlYaml, nil)

	tflog.Error(ctx, string(no_comments))

	return no_comments, nil
}

func resourceControlNodeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	bootstrap := d.Get("bootstrap").(bool)

	patched, diags := generateConfig(ctx, d)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}

	network := d.Get("dhcp_network_cidr").(string)
	mac := d.Get("macaddr").(string)

	dhcpIp, diags := lookupIP(ctx, network, mac)
	if diags != nil {
		tflog.Error(ctx, "Error looking up node IP")
		return diags
	}

	host := net.JoinHostPort(dhcpIp.String(), strconv.Itoa(talos_port))
	conn, diags := insecureConn(ctx, host)
	if diags != nil {
		return diags
	}
	defer conn.Close()
	client := machine.NewMachineServiceClient(conn)
	_, err := client.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: patched,
		Mode: machine.ApplyConfigurationRequest_Mode(machine.ApplyConfigurationRequest_REBOOT),
	})
	if err != nil {
		tflog.Error(ctx, "Error applying configuration")
		return diag.FromErr(err)
	}

	if bootstrap {
		ip := net.ParseIP(d.Get("bootstrap_ip").(string))
		if ip == nil {
			return diag.Errorf("Unable to parse bootstrap_ip")
		}
		host := net.JoinHostPort(ip.String(), strconv.Itoa(talos_port))
		input := generate.Input{}
		if err := json.Unmarshal([]byte(d.Get("base_config").(string)), &input); err != nil {
			return diag.FromErr(err)
		}

		conn, diags := secureConn(ctx, input, host)
		if diags != nil {
			return diags
		}
		defer conn.Close()
		client := machine.NewMachineServiceClient(conn)
		_, err = client.Bootstrap(ctx, &machine.BootstrapRequest{})
		if err != nil {
			tflog.Error(ctx, "Error getting Machine Configuration")
			return diag.FromErr(err)
		}
	}

	d.SetId(d.Get("name").(string))
	d.Set("patch", string(patched))

	return nil
}

func insecureConn(ctx context.Context, host string) (*grpc.ClientConn, diag.Diagnostics) {
	tlsConfig, diags := makeTlsConfig(generate.Certs{}, false)
	if diags != nil {
		return nil, diags
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)),
	}

	waitTillTalosMachineUp(ctx, tlsConfig, host, false)

	conn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		tflog.Error(ctx, "Error dailing talos.")
		return nil, diag.FromErr(err)
	}

	return conn, nil
}

func secureConn(ctx context.Context, input generate.Input, host string) (*grpc.ClientConn, diag.Diagnostics) {
	tlsConfig, diags := makeTlsConfig(*input.Certs, true)
	if diags != nil {
		return nil, diags
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)),
	}

	waitTillTalosMachineUp(ctx, tlsConfig, host, true)

	conn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		tflog.Error(ctx, "Error securely dailing talos.")
		return nil, diag.FromErr(err)
	}

	return conn, nil
}

func ipNetwork(ip net.IP, network net.IPNet) string {
	// Return interface IP followed by count of host identifier bits
	mask := network.Mask
	sum := 0
	for _, b := range mask {
		sum += bits.OnesCount8(uint8(b))
	}

	return ip.String() + "/" + strconv.Itoa(sum)
}

func resourceControlNodeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ip := net.ParseIP(d.Get("bootstrap_ip").(string))
	if ip == nil {
		return diag.Errorf("Unable to parse IP address from \"bootstrap_ip\", passed \"%s\",got \"%s\"", d.Get("bootstrap_ip").(string), ip.String())
	}
	host := net.JoinHostPort(ip.String(), strconv.Itoa(talos_port))

	input := generate.Input{}
	if err := json.Unmarshal([]byte(d.Get("base_config").(string)), &input); err != nil {
		return diag.FromErr(err)
	}

	conn, diags := secureConn(ctx, input, host)
	if diags != nil {
		return diags
	}

	defer conn.Close()
	client := resource.NewResourceServiceClient(conn)
	resp, err := client.Get(ctx, &resource.GetRequest{
		Type:      "MachineConfig",
		Namespace: "config",
		Id:        "v1alpha1",
	})
	if err != nil {
		tflog.Error(ctx, "Error getting Machine Configuration")
		return diag.FromErr(err)
	}

	if len(resp.Messages) < 1 {
		return diag.Errorf("Invalid message count recieved. Expected > 1 but got %d", len(resp.Messages))
	}

	conf := v1alpha1.Config{}
	err = yaml.Unmarshal(resp.Messages[0].Resource.Spec.Yaml, &conf)
	if err != nil {
		return diag.FromErr(err)
	}

	confNameservers := conf.MachineConfig.MachineNetwork.NameServers
	nameservers := make([](interface{}), len(confNameservers))
	for i := range confNameservers {
		nameservers[i] = confNameservers[i]
	}

	// Assume one regular interface and one wireguard interface. These will eventually be seperate types in terraform
	//talosInterfaces := conf.MachineConfig.MachineNetwork.NetworkInterfaces
	d.SetId(d.Get("name").(string))
	d.Set("name", conf.MachineConfig.MachineNetwork.NetworkHostname)
	d.Set("install_disk", conf.MachineConfig.MachineInstall.InstallDisk)
	d.Set("talos_image", conf.MachineConfig.MachineInstall.InstallImage)

	// Seperate wireguard and traditional interfaces
	/*
		wireguard := []*v1alpha1.Device{}
		networks := []*v1alpha1.Device{}
		netdevs := conf.MachineConfig.MachineNetwork.NetworkInterfaces
		for _, netdev := range netdevs {
			if netdev.DeviceWireguardConfig != nil {
				wireguard = append(wireguard, netdev)
			} else {
				networks = append(networks, netdev)
			}
		}

		d.Set("gateway", networks[0].DeviceRoutes[0].RouteGateway)
		d.Set("ip", networks[0].DeviceAddresses[0])
		d.Set("nameservers", nameservers)

		if len(talosInterfaces) > 1 {
			d.Set("wg_address", wireguard[0].DeviceAddresses[0])
			d.Set("wg_allowed_ips", wireguard[0].DeviceWireguardConfig.WireguardPeers[0].WireguardAllowedIPs[0])
			d.Set("wg_endpoint", wireguard[0].DeviceWireguardConfig.WireguardPeers[0].WireguardEndpoint)
			d.Set("wg_public_key", wireguard[0].DeviceWireguardConfig.WireguardPeers[0].WireguardPublicKey)
			d.Set("wf_private_key", wireguard[0].DeviceWireguardConfig.WireguardPrivateKey)
		}

		d.Set("local_api_proxy_port", conf.ClusterConfig.APIServerConfig.ExtraArgsConfig["secure-port"])
	*/
	return nil
}

func resourceControlNodeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	patched, diags := generateConfig(ctx, d)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}

	ip := net.ParseIP(d.Get("bootstrap_ip").(string))
	if ip == nil {
		return diag.Errorf("parsing IP %s", ip.String())
	}
	host := net.JoinHostPort(ip.String(), strconv.Itoa(talos_port))
	input := generate.Input{}
	if err := json.Unmarshal([]byte(d.Get("base_config").(string)), &input); err != nil {
		return diag.FromErr(err)
	}

	conn, diags := secureConn(ctx, input, host)
	if diags != nil {
		return diags
	}
	defer conn.Close()

	client := machine.NewMachineServiceClient(conn)
	resp, err := client.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: []byte(patched),
		Mode: machine.ApplyConfigurationRequest_Mode(machine.ApplyConfigurationRequest_AUTO),
	})
	if err != nil {
		tflog.Error(ctx, "Error applying configuration")
		return diag.FromErr(err)
	}
	log.Printf(resp.String())

	return nil
}

func resourceControlNodeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ip := net.ParseIP(d.Get("bootstrap_ip").(string))
	if ip == nil {
		return diag.Errorf("Invalid IP got %s", d.Get("bootstrap_ip").(string))
	}
	host := net.JoinHostPort(ip.String(), strconv.Itoa(talos_port))

	input := generate.Input{}
	if err := json.Unmarshal([]byte(d.Get("base_config").(string)), &input); err != nil {
		return diag.FromErr(err)
	}

	conn, diags := secureConn(ctx, input, host)
	if diags != nil {
		return diags
	}
	defer conn.Close()

	client := machine.NewMachineServiceClient(conn)
	resp, err := client.Reset(ctx, &machine.ResetRequest{
		Graceful: false,
		Reboot:   true,
	})
	if err != nil {
		tflog.Error(ctx, "Error resetting machine")
		return diag.FromErr(err)
	}
	log.Printf(resp.String())

	return nil
}

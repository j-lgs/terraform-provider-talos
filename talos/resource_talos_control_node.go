package talos

import (
	"bufio"
	"bytes"
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
	"text/template"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/yaml.v2"

	"crypto/tls"
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/api/resource"
	"github.com/talos-systems/talos/pkg/machinery/config/configpatcher"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	machinetype "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cluster_proxy_args": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"local_apiserver_port": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			// Network args
			"interface": &networkInterfaceSchema,

			"registry_mirrors": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateEndpoint,
				},
			},
			"nameservers": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateEndpoint,
				},
			},

			// Kubelet args
			"kubelet_extra_mount": &kubeletExtraMountSchema,
			"kubelet_extra_args": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// Cert args
			"cert_sans": {
				Type:     schema.TypeList,
				MinItems: 0,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIP,
				},
			},

			// Machinefiles
			"file": {
				Type:     schema.TypeList,
				MinItems: 0,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:     schema.TypeString,
							Required: true,
						},
						"permissions": {
							Type:     schema.TypeInt,
							Required: true,
							ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
								v := value.(int)
								if v < 0 {
									errs = append(errs, fmt.Errorf("Persistent keepalive interval must be a positive integer, got %d", v))
								}
								return
							},
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"op": {
							Type:     schema.TypeString,
							Required: true,
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

			// pods
			"pod": {
				Type:     schema.TypeList,
				MinItems: 0,
				Optional: true,
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

			// System args
			"sysctls": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// Cluster bootstrap configuration
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
		tflog.Error(ctx, "a")
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

func genKeypair(wgIp net.IP, wgPrivateKey string) (string, string, diag.Diagnostics) {
	// generate wireguard keypair
	pubkey := ""
	privkey := ""

	if wgIp.String() != "" {
		pk := wgPrivateKey
		if pk == "" {
			pk, err := wgtypes.GeneratePrivateKey()
			if err != nil {
				return "", "", diag.FromErr(err)
			}
			privkey = pk.String()
			pubkey = pk.PublicKey().String()
		} else {
			pk, err := wgtypes.ParseKey(pk)
			if err != nil {
				return "", "", diag.FromErr(err)
			}
			privkey = pk.String()
			pubkey = pk.PublicKey().String()
		}
	}

	return pubkey, privkey, nil
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

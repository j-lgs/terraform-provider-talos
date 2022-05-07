package talos

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math/bits"
	"net"
	"os"
	"os/exec"
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

func validateDomain(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("Node name must be a lowercase RFC 1123 subdomain, got %s", v))
	}
	return
}

func validateMAC(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, err := net.ParseMAC(v); err != nil {
		errs = append(errs, fmt.Errorf("Must provide a valid MAC address, got %s, error %s", v, err.Error()))
	}
	return
}

func validateCIDR(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, _, err := net.ParseCIDR(v); err != nil {
		errs = append(errs, fmt.Errorf("Must provide a valid CIDR IP address, got %s, error %s", v, err.Error()))
	}
	return
}

func validateIP(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if net.ParseIP(v) == nil {
		errs = append(errs, fmt.Errorf("Must provide a valid IP address, got %s", v))
	}
	return
}

func validateHost(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*:[0-9]{2,}`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("Node name must be a lowercase RFC 1123 subdomain with a port appended, seperated by \":\", got %s", v))
	}
	return
}

func validateImage(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[^/]+\.[^/.]+/([^/.]+/)?[^/.]+(:.+)?`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("Node name must be a valid container image, got %s", v))
	}
	return
}

func validateState(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	switch v {
	case
		"MASTER",
		"BACKUP":
	default:
		errs = append(errs, fmt.Errorf("Invalid keepalived node state, expected one of MASTER, BACKUP, got %s", v))
	}

	return
}

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
			},

			"install_disk": {
				Type:     schema.TypeString,
				Required: true,
			},
			"talos_image": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateImage,
			},

			"macaddr": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateMAC,
			},
			"ip": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDR,
			},
			"dhcp_network_cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDR,
			},
			"gateway": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateIP,
			},
			"nameservers": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateDomain,
				},
			},
			"peers": {
				Type:     schema.TypeList,
				MinItems: 0,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIP,
				},
			},

			// Cluster bootstrap configuration
			"bootstrap": {
				Type:     schema.TypeBool,
				Required: true,
			},

			// Wireguard optionals TODO make into typeset
			"wg_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCIDR,
			},
			"wg_allowed_ips": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCIDR,
			},
			"wg_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateHost,
			},
			"wg_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wg_private_key": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},

			// Haproxy wireguard ingress optionals
			"ingress_port": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  8080,
			},
			"ingress_ssl_port": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  8443,
			},
			"ingress_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateIP,
			},

			// Load balancing API proxy optionals
			"api_proxy_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateIP,
			},
			"api_proxy_port": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  6443,
			},
			"local_api_proxy_port": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  443,
			},

			// Shared IP optionals
			"router_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vrid": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "11",
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateState,
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"vip_pass": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			// Container registry optionals
			"registry_ip": {
				Type:         schema.TypeString,
				Optional:     true,
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

			// Container images
			"haproxy_image": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "haproxy:2.4.14",
				ValidateFunc: validateImage,
			},
			"keepalived_image": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "osixia/keepalived:1.3.5-1",
				ValidateFunc: validateImage,
			},
		},
	}
}

type ControlNodeSpec struct {
	Name string

	IP          string
	IPNetwork   string
	Hostname    string
	Gateway     string
	Nameservers []string
	Peers       []string

	WgIP         string
	WgAddress    string
	WgInterface  string
	WgAllowedIPs string
	WgEndpoint   string
	WgPublicKey  string
	WgPrivateKey string

	IngressPort    int
	IngressSSLPort int
	IngressIP      string

	RouterID string
	VRID     string
	State    string
	Priority int
	VIPPass  string

	APIProxyIP        string
	APIProxyPort      int
	LocalAPIProxyPort int

	RegistryIP string

	KeepalivedImage string
	HaproxyImage    string
}

func checkArp(mac net.HardwareAddr) (net.IP, diag.Diagnostics) {
	arp, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, diag.Errorf("%s\n", err)
	}
	defer arp.Close()

	scanner := bufio.NewScanner(arp)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if f[3] == mac.String() {
			return net.ParseIP(f[0]), nil
		}
	}

	return nil, nil
}

func lookupIP(ctx context.Context, network *net.IPNet, mac net.HardwareAddr) (net.IP, diag.Diagnostics) {
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
			err := exec.CommandContext(ctx, "nmap", "-sP", network.String()).Run()
			if err != nil {
				return nil, diag.Errorf("%s\n", err)
			}
			if ip, diags = checkArp(mac); diags != nil {
				return nil, diags
			}
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

func generatePatched(ctx context.Context, d *schema.ResourceData, config []byte) (string, diag.Diagnostics) {
	nameservers := []string{}
	for _, ns := range d.Get("nameservers").([]interface{}) {
		nameservers = append(nameservers, ns.(string))
	}

	var t *template.Template

	funcMap := template.FuncMap{
		"templateFilesValue": func(name string, op string, path string, permissions int, data interface{}) string {
			type ContentSpec struct {
				Content     string `json:"content"`
				Op          string `json:"op"`
				Path        string `json:"path"`
				Permissions int    `json:"permissions"`
			}

			buffer := &bytes.Buffer{}
			if err := t.ExecuteTemplate(buffer, name, data); err != nil {
				panic(err)
			}

			bytes, err := json.Marshal(ContentSpec{buffer.String(), op, path, permissions})
			if err != nil {
				panic(err)
			}

			if !json.Valid(bytes) {
				panic("invalid JSON")
			}

			return string(bytes)
		},
	}

	// template controlplane patches
	t = template.Must(template.New("controlPlane").Funcs(funcMap).Parse(templateControl()))
	t = template.Must(t.Parse(templateHaproxy()))
	t = template.Must(t.Parse(templateKeepalived()))
	t = template.Must(t.Parse(templateAPICheck()))

	peerList := d.Get("peers").([]interface{})
	peers := []string{}
	for _, peer := range peerList {
		peers = append(peers, peer.(string))
	}

	ip, network, err := net.ParseCIDR(d.Get("ip").(string))
	if err != nil {
		return "", diag.FromErr(err)
	}

	wgAddress := d.Get("wg_address").(string)
	wgAddress_ := ""
	wgIp, wgNetwork := net.IP{}, &net.IPNet{}
	if wgAddress == "" {
		wgAddress_ = ""
	} else {
		wgIp, wgNetwork, err = net.ParseCIDR(wgAddress)
		if err != nil {
			return "", diag.FromErr(err)
		}
		wgAddress = ipNetwork(wgIp, *wgNetwork)
	}

	buffer := new(strings.Builder)
	err = t.ExecuteTemplate(buffer, "controlPlane", ControlNodeSpec{
		Name:      d.Get("name").(string),
		IP:        ip.String(),
		IPNetwork: ipNetwork(ip, *network),

		Hostname:    d.Get("name").(string),
		Gateway:     d.Get("gateway").(string),
		Nameservers: nameservers,
		Peers:       peers,

		WgIP:         wgIp.String(),
		WgAddress:    wgAddress_,
		WgInterface:  "wg0",
		WgAllowedIPs: d.Get("wg_allowed_ips").(string),
		WgEndpoint:   d.Get("wg_endpoint").(string),
		WgPublicKey:  d.Get("wg_public_key").(string),
		WgPrivateKey: d.Get("wg_private_key").(string),

		IngressPort:    d.Get("ingress_port").(int),
		IngressSSLPort: d.Get("ingress_ssl_port").(int),
		IngressIP:      d.Get("ingress_ip").(string),

		RouterID: d.Get("router_id").(string),
		VRID:     d.Get("vrid").(string),
		State:    d.Get("state").(string),
		Priority: d.Get("priority").(int),
		VIPPass:  d.Get("vip_pass").(string),

		APIProxyIP:        d.Get("api_proxy_ip").(string),
		APIProxyPort:      d.Get("api_proxy_port").(int),
		LocalAPIProxyPort: d.Get("local_api_proxy_port").(int),

		RegistryIP: d.Get("registry_ip").(string),

		KeepalivedImage: d.Get("keepalived_image").(string),
		HaproxyImage:    d.Get("haproxy_image").(string),
	})
	if err != nil {
		tflog.Error(ctx, "Error running controlplane template.")
		return "", diag.FromErr(err)
	}

	jsonpatch, err := jsonpatch.DecodePatch([]byte(buffer.String()))
	if err != nil {
		tflog.Error(ctx, "Error decoding jsonpatch: "+buffer.String())
		return "", diag.FromErr(err)
	}

	patched, err := configpatcher.JSON6902(config, jsonpatch)
	if err != nil {
		tflog.Error(ctx, "Error attempting applying jsonpatch: "+buffer.String())
		return "", diag.FromErr(err)
	}

	return string(patched), nil
}

func generateConfig(ctx context.Context, d *schema.ResourceData) ([]byte, diag.Diagnostics) {
	disk := d.Get("install_disk").(string)
	image := d.Get("talos_image").(string)

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

	controlCfg.MachineConfig.MachineInstall.InstallDisk = disk
	controlCfg.MachineConfig.MachineInstall.InstallImage = image
	var controlYaml []byte

	controlYaml, err = controlCfg.Bytes()
	if err != nil {
		log.Fatalf("failed to generate config" + err.Error())
		return nil, diag.FromErr(err)
	}

	return controlYaml, nil
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

	wgIp, _, err := net.ParseCIDR(d.Get("wg_address").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	pub, priv, diags := genKeypair(wgIp, d.Get("wg_private_key").(string))
	if diags != nil {
		return diags
	}
	d.Set("wg_public_key", pub)
	d.Set("wg_private_key", priv)

	cfg, diags := generateConfig(ctx, d)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}

	patched, diags := generatePatched(ctx, d, cfg)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}

	_, network, err := net.ParseCIDR(d.Get("dhcp_network_cidr").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	mac, err := net.ParseMAC(d.Get("macaddr").(string))
	if err != nil {
		return diag.FromErr(err)
	}
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
	_, err = client.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: []byte(patched),
		Mode: machine.ApplyConfigurationRequest_Mode(machine.ApplyConfigurationRequest_REBOOT),
	})
	if err != nil {
		tflog.Error(ctx, "Error applying configuration")
		return diag.FromErr(err)
	}

	if bootstrap {
		ip, _, err := net.ParseCIDR(d.Get("ip").(string))
		if err != nil {
			return diag.FromErr(err)
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
	d.Set("wg_public_key", pub)
	d.Set("wg_private_key", priv)
	d.Set("patch", patched)

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

	ip, _, err := net.ParseCIDR(d.Get("ip").(string))
	if err != nil {
		return diag.FromErr(err)
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
	talosInterfaces := conf.MachineConfig.MachineNetwork.NetworkInterfaces
	d.SetId(d.Get("name").(string))
	d.Set("name", conf.MachineConfig.MachineNetwork.NetworkHostname)
	d.Set("install_disk", conf.MachineConfig.MachineInstall.InstallDisk)
	d.Set("talos_image", conf.MachineConfig.MachineInstall.InstallImage)

	// Seperate wireguard and traditional interfaces
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

	return nil
}

func resourceControlNodeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cfg, diags := generateConfig(ctx, d)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}

	patched, diags := generatePatched(ctx, d, cfg)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched configuration")
		return diags
	}

	ip, _, err := net.ParseCIDR(d.Get("ip").(string))
	if err != nil {
		tflog.Error(ctx, "parsing IP CIDR")
		return diag.FromErr(err)
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
	ip, _, err := net.ParseCIDR(d.Get("ip").(string))
	if err != nil {
		return diag.FromErr(err)
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

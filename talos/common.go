package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/opencontainers/runtime-spec/specs-go"
	talosx509 "github.com/talos-systems/crypto/x509"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/api/resource"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	machinetype "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Interface defining methods used to move data to and from talos and terraform.
//lint:ignore U1000 type exists just to define the interface and isn't used by itself.
type planToAPI interface {
	Data() (interface{}, error)
	Read(interface{}) error
}

// VolumeMount Describes extra volume mounts for controlplane static pods.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#volumemountconfig for more information.
type VolumeMount struct {
	HostPath  types.String `tfsdk:"host_path"`
	MountPath types.String `tfsdk:"mount_path"`
	Readonly  types.Bool   `tfsdk:"readonly"`
}

// VolumeMountSchema Describes extra volume mounts for controlplane static pods.
var VolumeMountSchema tfsdk.Schema = tfsdk.Schema{
	MarkdownDescription: "Describes extra volume mouns for controlplane static pods.",
	Attributes: map[string]tfsdk.Attribute{
		"host_path": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Path on the host.",
			// TODO validate it is a well formed path
		},
		"mount_path": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Path in the container.",
			// TODO validate it is a well formed path
		},
		"readonly": {
			Type:                types.BoolType,
			Optional:            true,
			MarkdownDescription: "Mount the volume read only.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (mount VolumeMount) Data() (interface{}, error) {
	vol := v1alpha1.VolumeMountConfig{
		VolumeHostPath:  mount.HostPath.Value,
		VolumeMountPath: mount.MountPath.Value,
	}
	if !mount.Readonly.Null {
		vol.VolumeReadOnly = mount.Readonly.Value
	}

	return vol, nil
}

// Read copies data from talos types to terraform state types.
func (mount *VolumeMount) Read(vol interface{}) error {
	volume := vol.(v1alpha1.VolumeMountConfig)

	mount.HostPath = types.String{Value: volume.VolumeHostPath}
	mount.MountPath = types.String{Value: volume.VolumeMountPath}
	mount.Readonly = types.Bool{Value: volume.VolumeReadOnly}

	return nil
}

// ExtraMount wraps the OCI mount specification.
// Refer to https://github.com/opencontainers/runtime-spec/blob/main/config.md#mounts for more information.
type ExtraMount struct {
	Destination types.String   `tfsdk:"destination"`
	Type        types.String   `tfsdk:"type"`
	Source      types.String   `tfsdk:"source"`
	Options     []types.String `tfsdk:"options"`
}

// ExtraMountSchema wraps the OCI mount specification.
var ExtraMountSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Wraps the OCI Mount specification.",
	Attributes: map[string]tfsdk.Attribute{
		"destination": {
			Type:        types.StringType,
			Required:    true,
			Description: "Destination of mount point: path inside container. This value MUST be an absolute path.",
		},
		"type": {
			Type:        types.StringType,
			Optional:    true,
			Description: "The type of the filesystem to be mounted.",
			//			ValidateFunc:
		},
		"source": {
			Type:        types.StringType,
			Required:    true,
			Description: "A device name, but can also be a file or directory name for bind mounts or a dummy. Path values for bind mounts are either absolute or relative to the bundle. A mount is a bind mount if it has either bind or rbind in the options.",
			// TODO: Add singleton validator. IsValid(f),
			//Validators: []tfsdk.AttributeValidator{
			//	AllElemsValid(IsValidPath),
			//},
		},
		"options": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Mount options of the filesystem to be used.",
			// TODO: Replace validator with proper one for mount options.
			//Validators: []tfsdk.AttributeValidator{
			//	AllElemsValid(IsValidPath),
			//},
		},
	},
}

// Data copies data from terraform state types to talos types.
func (mount ExtraMount) Data() (interface{}, error) {
	extraMount := v1alpha1.ExtraMount{
		Mount: specs.Mount{
			Destination: mount.Destination.Value,
			Source:      mount.Source.Value,
			Type:        mount.Type.Value,
		},
	}

	for _, opt := range mount.Options {
		extraMount.Options = append(extraMount.Options, opt.Value)
	}

	return extraMount, nil
}

// Read copies data from talos types to terraform state types.
func (mount *ExtraMount) Read(mnt interface{}) error {
	talosMount := mnt.(v1alpha1.ExtraMount)
	mount.Destination = types.String{Value: talosMount.Destination}
	mount.Source = types.String{Value: talosMount.Source}

	if talosMount.Type != "" {
		mount.Type = types.String{Value: talosMount.Type}
	}

	for _, opt := range talosMount.Options {
		mount.Options = append(mount.Options, types.String{Value: opt})
	}

	return nil
}

// KubeletConfig represents the kubelet's config values.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#kubeletconfig for more information.
type KubeletConfig struct {
	Image              types.String            `tfsdk:"image"`
	ClusterDNS         []types.String          `tfsdk:"cluster_dns"`
	ExtraArgs          map[string]types.String `tfsdk:"extra_args"`
	ExtraMounts        []ExtraMount            `tfsdk:"extra_mount"`
	ExtraConfig        types.String            `tfsdk:"extra_config"`
	RegisterWithFQDN   types.Bool              `tfsdk:"register_with_fqdn"`
	NodeIPValidSubnets []types.String          `tfsdk:"node_ip_valid_subnets"`
}

// KubeletConfigSchema represents the kubelet's config values.
var KubeletConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kubelet's config values.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:        types.StringType,
			Required:    true,
			Description: "An optional reference to an alternative kubelet image.",
			//			ValidateFunc: validateImage,
		},
		// TODO: Add validator for IP
		"cluster_dns": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "An optional reference to an alternative kubelet clusterDNS ip list.",
			Optional:    true,
		},
		"extra_args": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Used to provide additional flags to the kubelet.",
		},
		"extra_mount": {
			Optional:    true,
			Attributes:  tfsdk.ListNestedAttributes(ExtraMountSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			Description: ExtraMountSchema.Description,
		},
		// TODO Add yaml validation function
		"extra_config": {
			Type:        types.StringType,
			Optional:    true,
			Description: "The extraConfig field is used to provide kubelet configuration overrides. Must be valid YAML",
		},
		"register_with_fqdn": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Used to force kubelet to use the node FQDN for registration. This is required in clouds like AWS.",
		},
		// TODO: Add validator
		"node_ip_valid_subnets": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "The validSubnets field configures the networks to pick kubelet node IP from.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (kubelet KubeletConfig) Data() (interface{}, error) {
	talosKubelet := &v1alpha1.KubeletConfig{}
	if !kubelet.Image.Null {
		talosKubelet.KubeletImage = kubelet.Image.Value
	}
	if !kubelet.RegisterWithFQDN.Null {
		talosKubelet.KubeletRegisterWithFQDN = kubelet.RegisterWithFQDN.Value
	}
	if !kubelet.ExtraConfig.Null {
		var conf v1alpha1.Unstructured
		if err := yaml.Unmarshal([]byte(kubelet.ExtraConfig.Value), &conf); err != nil {
			return nil, nil
		}

		talosKubelet.KubeletExtraConfig = conf
	}
	for _, dns := range kubelet.ClusterDNS {
		talosKubelet.KubeletClusterDNS = append(talosKubelet.KubeletClusterDNS, dns.Value)
	}
	if len(kubelet.ExtraArgs) > 0 {
		talosKubelet.KubeletExtraArgs = map[string]string{}
	}
	for k, arg := range kubelet.ExtraArgs {
		talosKubelet.KubeletExtraArgs[k] = arg.Value
	}
	for _, mount := range kubelet.ExtraMounts {
		m, err := mount.Data()
		if err != nil {
			return nil, err
		}
		talosKubelet.KubeletExtraMounts = append(talosKubelet.KubeletExtraMounts, m.(v1alpha1.ExtraMount))
	}
	if len(kubelet.NodeIPValidSubnets) > 0 {
		talosKubelet.KubeletNodeIP = v1alpha1.KubeletNodeIPConfig{}
		for _, subnet := range kubelet.NodeIPValidSubnets {
			talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets =
				append(talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets, subnet.Value)
		}
	}
	return talosKubelet, nil
}

// Read copies data from talos types to terraform state types.
func (kubelet *KubeletConfig) Read(talosData interface{}) error {
	talosKubelet := talosData.(*v1alpha1.KubeletConfig)
	if talosKubelet.KubeletImage != "" {
		kubelet.Image = types.String{Value: talosKubelet.KubeletImage}
	}

	kubelet.RegisterWithFQDN = types.Bool{Value: talosKubelet.KubeletRegisterWithFQDN}

	if !reflect.DeepEqual(talosKubelet.KubeletExtraConfig.Object, map[string]interface{}{}) {
		bytes, err := yaml.Marshal(&talosKubelet.KubeletExtraConfig)
		if err != nil {
			return err
		}
		kubelet.ExtraConfig = types.String{Value: string(bytes)}
	}

	for _, dns := range talosKubelet.KubeletClusterDNS {
		kubelet.ClusterDNS = append(kubelet.ClusterDNS, types.String{Value: dns})
	}

	for _, mount := range talosKubelet.KubeletExtraMounts {
		extraMount := ExtraMount{}
		err := extraMount.Read(mount)
		if err != nil {
			return err
		}
		kubelet.ExtraMounts = append(kubelet.ExtraMounts, extraMount)
	}

	if !reflect.DeepEqual(talosKubelet.KubeletNodeIP, v1alpha1.KubeletNodeIPConfig{}) {
		for _, subnet := range talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets {
			kubelet.NodeIPValidSubnets = append(kubelet.NodeIPValidSubnets, types.String{Value: subnet})
		}
	}

	return nil
}

// Registry represents the image pull options.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#registriesconfig for more information.
type Registry struct {
	Mirrors map[string][]types.String `tfsdk:"mirrors"`
	Configs map[string]RegistryConfig `tfsdk:"configs"`
}

// RegistrySchema represents the image pull options.
var RegistrySchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the image pull options.",
	Attributes: map[string]tfsdk.Attribute{
		"mirrors": {
			Optional: true,
			Type: types.MapType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			Description: "Specifies mirror configuration for each registry.",
		},
		"configs": {
			Optional:    true,
			Description: RegistryConfigSchema.Description,
			Attributes:  tfsdk.MapNestedAttributes(RegistryConfigSchema.Attributes, tfsdk.MapNestedAttributesOptions{}),
		},
	},
}

// Data copies data from terraform state types to talos types.
func (registry Registry) Data() (interface{}, error) {
	regs := &v1alpha1.RegistriesConfig{}

	regs.RegistryMirrors = map[string]*v1alpha1.RegistryMirrorConfig{}
	for registry, endpoints := range registry.Mirrors {
		regs.RegistryMirrors[registry] = &v1alpha1.RegistryMirrorConfig{}
		for _, endpoint := range endpoints {
			regs.RegistryMirrors[registry].MirrorEndpoints = append(regs.RegistryMirrors[registry].MirrorEndpoints, endpoint.Value)
		}
	}

	if registry.Configs != nil {
		regs.RegistryConfig = map[string]*v1alpha1.RegistryConfig{}
		for registry, conf := range registry.Configs {
			config, err := conf.Data()
			if err != nil {
				return nil, err
			}
			regs.RegistryConfig[registry] = config.(*v1alpha1.RegistryConfig)

		}
	}

	return regs, nil
}

// RegistryConfig specifies TLS & auth configuration for HTTPS image registries. The meaning of each
// auth_field is the same with the corresponding field in .docker/config.json."
type RegistryConfig struct {
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Auth               types.String `tfsdk:"auth"`
	IdentityToken      types.String `tfsdk:"identity_token"`
	ClientCRT          types.String `tfsdk:"client_identity_crt"`
	ClientKey          types.String `tfsdk:"client_identity_key"`
	CA                 types.String `tfsdk:"ca"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

// RegistryConfigSchema specifies TLS & auth configuration for HTTPS image registries. The meaning of each
// auth_field is the same with the corresponding field in .docker/config.json."
var RegistryConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: `Specifies TLS & auth configuration for HTTPS image registries. The meaning of each auth_field is the same with the corresponding field in .docker/config.json.

Key description: The first segment of an image identifier, with ‘docker.io’ being default one. To catch any registry names not specified explicitly, use ‘*’.`,
	Attributes: map[string]tfsdk.Attribute{
		"username": {
			Type:        types.StringType,
			Optional:    true,
			Description: "Username for optional registry authentication.",
		},
		"password": {
			Type:        types.StringType,
			Optional:    true,
			Sensitive:   true,
			Description: "Password for optional registry authentication.",
		},
		"auth": {
			Type:        types.StringType,
			Optional:    true,
			Sensitive:   true,
			Description: "Auth for optional registry authentication.",
		},
		"identity_token": {
			Type:        types.StringType,
			Optional:    true,
			Sensitive:   true,
			Description: "Identity token for optional registry authentication.",
		},
		// It seems that when marshalled to yaml these values are automatically base64 encoded. Therefore we must ensure that it is
		// not base64 encoded.
		"client_identity_crt": {
			Type:        types.StringType,
			Optional:    true,
			Sensitive:   true,
			Description: "Enable mutual TLS authentication with the registry. Non base64 encoded client certificate.",
			// TODO: validate it's a correctly encoded PEM certificate and not valid base64
		},
		"client_identity_key": {
			Type:        types.StringType,
			Optional:    true,
			Sensitive:   true,
			Description: "Enable mutual TLS authentication with the registry. Non base64 encoded client key.",
			// TODO: validate it's a correctly encoded PEM key and not valid base64
		},
		"ca": {
			Type:        types.StringType,
			Optional:    true,
			Description: "CA registry certificate to add the list of trusted certificates. Non base64 encoded.",
			// TODO: Verify CA is base64 encoded
		},
		"insecure_skip_verify": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Skip TLS server certificate verification (not recommended)..",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (config RegistryConfig) Data() (interface{}, error) {
	conf := &v1alpha1.RegistryConfig{}

	conf.RegistryTLS = &v1alpha1.RegistryTLSConfig{}
	conf.RegistryAuth = &v1alpha1.RegistryAuthConfig{}
	if !config.ClientCRT.Null {
		conf.RegistryTLS.TLSClientIdentity = &talosx509.PEMEncodedCertificateAndKey{
			Crt: []byte(config.ClientCRT.Value),
			Key: []byte(config.ClientKey.Value),
		}
	}
	if !config.CA.Null {
		conf.RegistryTLS.TLSCA = []byte(config.CA.Value)
	}
	if !config.InsecureSkipVerify.Null {
		conf.RegistryTLS.TLSInsecureSkipVerify = config.InsecureSkipVerify.Value
	}

	if !config.Username.Null && !config.Password.Null {
		conf.RegistryAuth.RegistryUsername = config.Username.Value
		conf.RegistryAuth.RegistryPassword = config.Password.Value
	}

	if !config.Auth.Null && !config.IdentityToken.Null {
		conf.RegistryAuth.RegistryAuth = config.Auth.Value
		conf.RegistryAuth.RegistryIdentityToken = config.IdentityToken.Value
	}

	return conf, nil
}

type NetworkConfig struct {
	Hostname    types.String              `tfsdk:"hostname"`
	Devices     map[string]NetworkDevice  `tfsdk:"devices"`
	Nameservers []types.String            `tfsdk:"nameservers"`
	ExtraHosts  map[string][]types.String `tfsdk:"extra_hosts"`
	Kubespan    types.Bool                `tfsdk:"kubespan"`
}

var NetworkConfigSchema = tfsdk.Schema{
	MarkdownDescription: "Represents node network configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"hostname": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Used to statically set the hostname for the machine..",
		},
		"devices": {
			Optional:    true,
			Description: NetworkDeviceSchema.Description,
			Attributes:  tfsdk.MapNestedAttributes(NetworkDeviceSchema.Attributes, tfsdk.MapNestedAttributesOptions{}),
		},
		"nameservers": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Used to statically set the nameservers for the machine.",
		},
		"extra_hosts": {
			Type: types.MapType{
				ElemType: types.ListType{
					ElemType: types.StringType,
				},
			},
			Optional:    true,
			Description: "Allows for extra entries to be added to the `/etc/hosts` file.",
		},
		"kubespan": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Configures the KubeSpan wireguard network feature.",
		},
	},
}

// NetworkDevice describes a Talos Device configuration.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#device for more information.
// TODO: Add network device selector field for interfaces and support it throughout the provider.
type NetworkDevice struct {
	Addresses   []types.String `tfsdk:"addresses"`
	Routes      []Route        `tfsdk:"routes"`
	BondData    *BondData      `tfsdk:"bond"`
	VLANs       []VLAN         `tfsdk:"vlans"`
	MTU         types.Int64    `tfsdk:"mtu"`
	DHCP        types.Bool     `tfsdk:"dhcp"`
	DHCPOptions *DHCPOptions   `tfsdk:"dhcp_options"`
	Ignore      types.Bool     `tfsdk:"ignore"`
	Dummy       types.Bool     `tfsdk:"dummy"`
	Wireguard   *Wireguard     `tfsdk:"wireguard"`
	VIP         *VIP           `tfsdk:"vip"`
}

// NetworkDeviceSchema describes a Talos Device configuration.
var NetworkDeviceSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Describes a Talos network device configuration. The map's key is the interface name.",
	Attributes: map[string]tfsdk.Attribute{
		"addresses": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required:    true,
			Description: "A list of IP addresses for the interface.",
			// TODO Add field validation
		},

		"routes": {
			Optional:    true,
			Attributes:  tfsdk.ListNestedAttributes(RouteSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			Description: RouteSchema.Description,
		},
		// Broken in a way I cannot currently comprehend.
		// TODO Find a fix for this schema breaking terraform.
		"bond": {
			Optional:    true,
			Attributes:  tfsdk.SingleNestedAttributes(BondSchema.Attributes),
			Description: BondSchema.Description,
		},
		"vlans": {
			Optional:    true,
			Attributes:  tfsdk.ListNestedAttributes(VLANSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			Description: VLANSchema.Description,
		},

		"mtu": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "The interface’s MTU. If used in combination with DHCP, this will override any MTU settings returned from DHCP server.",
		},
		"dhcp": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Indicates if DHCP should be used to configure the interface.",
		},
		"ignore": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Indicates if the interface should be ignored (skips configuration).",
		},
		"dummy": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Indicates if the interface is a dummy interface..",
		},

		"dhcp_options": {
			Optional:    true,
			Attributes:  tfsdk.SingleNestedAttributes(WireguardSchema.Attributes),
			Description: WireguardSchema.Description,
		},
		"wireguard": {
			Optional:    true,
			Attributes:  tfsdk.SingleNestedAttributes(WireguardSchema.Attributes),
			Description: WireguardSchema.Description,
		},
		"vip": {
			Optional:    true,
			Attributes:  tfsdk.SingleNestedAttributes(VIPSchema.Attributes),
			Description: VIPSchema.Description,
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planDevice NetworkDevice) Data() (interface{}, error) {
	device := &v1alpha1.Device{}

	for _, address := range planDevice.Addresses {
		device.DeviceAddresses = append(device.DeviceAddresses, address.Value)
	}

	for _, planRoute := range planDevice.Routes {
		route, err := planRoute.Data()
		if err != nil {
			return &v1alpha1.Device{}, err
		}
		device.DeviceRoutes = append(device.DeviceRoutes, route.(*v1alpha1.Route))
	}

	if planDevice.BondData != nil {
		bond, err := planDevice.BondData.Data()
		if err != nil {
			return nil, err
		}
		device.DeviceBond = bond.(*v1alpha1.Bond)
	}

	if !planDevice.DHCP.Null {
		device.DeviceDHCP = planDevice.DHCP.Value
	}

	if planDevice.DHCPOptions != nil {
		dhcpopts, err := planDevice.DHCPOptions.Data()
		if err != nil {
			return &v1alpha1.Device{}, err
		}
		device.DeviceDHCPOptions = dhcpopts.(*v1alpha1.DHCPOptions)
	}

	for _, planVLAN := range planDevice.VLANs {
		vlan, err := planVLAN.Data()
		if err != nil {
			return &v1alpha1.Device{}, err
		}
		device.DeviceVlans = append(device.DeviceVlans, vlan.(*v1alpha1.Vlan))
	}

	if !planDevice.MTU.Null {
		device.DeviceMTU = int(planDevice.MTU.Value)
	}

	if !planDevice.DHCP.Null {
		device.DeviceDHCP = planDevice.DHCP.Value
	}

	if !planDevice.Ignore.Null {
		device.DeviceIgnore = planDevice.Ignore.Value
	}

	if !planDevice.Dummy.Null {
		device.DeviceDummy = planDevice.Dummy.Value
	}
	if planDevice.Wireguard != nil {
		wireguard, err := planDevice.Wireguard.Data()
		if err != nil {
			return v1alpha1.Device{}, err
		}
		device.DeviceWireguardConfig = wireguard.(*v1alpha1.DeviceWireguardConfig)
	}
	if planDevice.VIP != nil {
		vip, err := planDevice.VIP.Data()
		if err != nil {
			return v1alpha1.Device{}, err
		}
		device.DeviceVIPConfig = vip.(*v1alpha1.DeviceVIPConfig)
	}

	return device, nil
}

// BondData contains the various options for configuring a bonded interface.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#bond for more information.
type BondData struct {
	Interfaces      []types.String `tfsdk:"interfaces"`
	ARPIPTarget     []types.String `tfsdk:"arp_ip_target"`
	Mode            types.String   `tfsdk:"mode"`
	XmitHashPolicy  types.String   `tfsdk:"xmit_hash_policy"`
	LacpRate        types.String   `tfsdk:"lacp_rate"`
	AdActorSystem   types.String   `tfsdk:"ad_actor_system"`
	ArpValidate     types.String   `tfsdk:"arp_validate"`
	ArpAllTargets   types.String   `tfsdk:"arp_all_targets"`
	Primary         types.String   `tfsdk:"primary"`
	PrimaryReselect types.String   `tfsdk:"primary_reselect"`
	FailoverMac     types.String   `tfsdk:"failover_mac"`
	AdSelect        types.String   `tfsdk:"ad_select"`
	MiiMon          types.Int64    `tfsdk:"mii_mon"`
	UpDelay         types.Int64    `tfsdk:"up_delay"`
	DownDelay       types.Int64    `tfsdk:"down_delay"`
	ArpInterval     types.Int64    `tfsdk:"arp_interval"`
	ResendIgmp      types.Int64    `tfsdk:"resend_igmp"`
	MinLinks        types.Int64    `tfsdk:"min_links"`
	LpInterval      types.Int64    `tfsdk:"lp_interval"`
	PacketsPerSlave types.Int64    `tfsdk:"packets_per_slave"`
	NumPeerNotif    types.Int64    `tfsdk:"num_peer_notif"`
	TlbDynamicLb    types.Int64    `tfsdk:"tlb_dynamic_lb"`
	AllSlavesActive types.Int64    `tfsdk:"all_slaves_active"`
	UseCarrier      types.Bool     `tfsdk:"use_carrier"`
	AdActorSysPrio  types.Int64    `tfsdk:"ad_actor_sys_prio"`
	AdUserPortKey   types.Int64    `tfsdk:"ad_user_port_key"`
	PeerNotifyDelay types.Int64    `tfsdk:"peer_notify_delay"`
}

// BondSchema contains the various options for configuring a bonded interface.
// TODO identify why including this schema breaks the provider.
var BondSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Contains the various options for configuring a bonded interface.",
	Attributes: map[string]tfsdk.Attribute{
		"interfaces": {
			Required: true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
		},
		"arp_ip_target": {
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
		},
		"mode": {
			Type:        types.StringType,
			Required:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"xmit_hash_policy": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"lacp_rate": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"ad_actor_system": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"arp_validate": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"arp_all_targets": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"primary": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"primary_reselect": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"failover_mac": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"ad_select": {
			Type:        types.StringType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"mii_mon": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"up_delay": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"down_delay": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"arp_interval": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"resend_igmp": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"min_links": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"lp_interval": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"packets_per_slave": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
		"num_peer_notif": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.",
		},
		"tlb_dynamic_lb": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.",
		},
		"all_slaves_active": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.",
		},
		"use_carrier": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation.",
		},
		"ad_actor_sys_prio": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 16 bit unsigned int.",
		},
		"ad_user_port_key": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 16 bit unsigned int.",
		},
		"peer_notify_delay": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planBond BondData) Data() (interface{}, error) {
	bond := &v1alpha1.Bond{}
	for _, netInterface := range planBond.Interfaces {
		bond.BondInterfaces = append(bond.BondInterfaces, netInterface.Value)
	}
	for _, arpIPTarget := range planBond.ARPIPTarget {
		bond.BondARPIPTarget = append(bond.BondARPIPTarget, arpIPTarget.Value)
	}
	b := planBond
	bond.BondMode = b.Mode.Value

	if !b.XmitHashPolicy.Null {
		bond.BondHashPolicy = b.XmitHashPolicy.Value
	}
	if !b.LacpRate.Null {
		bond.BondLACPRate = b.LacpRate.Value
	}
	if !b.AdActorSystem.Null {
		bond.BondADActorSystem = b.AdActorSystem.Value
	}
	if !b.ArpValidate.Null {
		bond.BondARPValidate = b.ArpValidate.Value
	}
	if !b.ArpAllTargets.Null {
		bond.BondARPAllTargets = b.ArpAllTargets.Value
	}
	if !b.Primary.Null {
		bond.BondPrimary = b.Primary.Value
	}
	if !b.PrimaryReselect.Null {
		bond.BondPrimaryReselect = b.PrimaryReselect.Value
	}
	if !b.FailoverMac.Null {
		bond.BondFailOverMac = b.FailoverMac.Value
	}
	if !b.AdSelect.Null {
		bond.BondADSelect = b.AdSelect.Value
	}
	if !b.MiiMon.Null {
		bond.BondMIIMon = uint32(b.MiiMon.Value)
	}
	if !b.UpDelay.Null {
		bond.BondUpDelay = uint32(b.UpDelay.Value)
	}
	if !b.DownDelay.Null {
		bond.BondDownDelay = uint32(b.DownDelay.Value)
	}
	if !b.ArpInterval.Null {
		bond.BondARPInterval = uint32(b.ArpInterval.Value)
	}
	if !b.ResendIgmp.Null {
		bond.BondResendIGMP = uint32(b.ResendIgmp.Value)
	}
	if !b.MinLinks.Null {
		bond.BondMinLinks = uint32(b.MinLinks.Value)
	}
	if !b.LpInterval.Null {
		bond.BondLPInterval = uint32(b.LpInterval.Value)
	}
	if !b.PacketsPerSlave.Null {
		bond.BondPacketsPerSlave = uint32(b.PacketsPerSlave.Value)
	}
	if !b.NumPeerNotif.Null {
		bond.BondNumPeerNotif = uint8(b.NumPeerNotif.Value)
	}
	if !b.TlbDynamicLb.Null {
		bond.BondTLBDynamicLB = uint8(b.TlbDynamicLb.Value)
	}
	if !b.AllSlavesActive.Null {
		bond.BondAllSlavesActive = uint8(b.AllSlavesActive.Value)
	}
	if !b.UseCarrier.Null {
		bond.BondUseCarrier = &b.UseCarrier.Value
	}
	if !b.AdActorSysPrio.Null {
		bond.BondADActorSysPrio = uint16(b.AdActorSysPrio.Value)
	}
	if !b.AdUserPortKey.Null {
		bond.BondADUserPortKey = uint16(b.AdUserPortKey.Value)
	}
	if !b.PeerNotifyDelay.Null {
		bond.BondPeerNotifyDelay = uint32(b.PeerNotifyDelay.Value)
	}

	return bond, nil
}

// DHCPOptions specificies DHCP specific options.
type DHCPOptions struct {
	RouteMetric types.Int64 `tfsdk:"route_metric"`
	IPV4        types.Bool  `tfsdk:"ipv4"`
	IPV6        types.Bool  `tfsdk:"ipv6"`
}

// DHCPOptionsSchema specificies DHCP specific options.
var DHCPOptionsSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Specifies DHCP specific options.",
	Attributes: map[string]tfsdk.Attribute{
		"route_metric": {
			Type:        types.Int64Type,
			Required:    true,
			Description: "The priority of all routes received via DHCP. Must be castable to a uint32.",
		},
		"ipv4": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Enables DHCPv4 protocol for the interface.",
			// TODO: Set default to true
		},
		"ipv6": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Enables DHCPv6 protocol for the interface.",
		},
	},
}

// Data copies data from terraform state types to talos types.
// TODO add DHCPUIDv6 from struct, and add it to the Talos config documentation.
func (planDHCPOptions DHCPOptions) Data() (interface{}, error) {
	dhcpOptions := &v1alpha1.DHCPOptions{}

	if !planDHCPOptions.IPV4.Null {
		dhcpOptions.DHCPIPv4 = &planDHCPOptions.IPV4.Value
	}
	if !planDHCPOptions.IPV6.Null {
		dhcpOptions.DHCPIPv6 = &planDHCPOptions.IPV6.Value
	}
	if !planDHCPOptions.RouteMetric.Null {
		dhcpOptions.DHCPRouteMetric = uint32(planDHCPOptions.RouteMetric.Value)
	}

	return dhcpOptions, nil
}

// VLAN represents vlan settings for a network device.
type VLAN struct {
	Addresses []types.String `tfsdk:"addresses"`
	Routes    []Route        `tfsdk:"routes"`
	DHCP      types.Bool     `tfsdk:"dhcp"`
	VLANId    types.Int64    `tfsdk:"vlan_id"`
	MTU       types.Int64    `tfsdk:"mtu"`
	VIP       *VIP           `tfsdk:"vip"`
}

// VLANSchema represents vlan settings for a network device.
var VLANSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents vlan settings for a device.",
	Attributes: map[string]tfsdk.Attribute{
		"addresses": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "A list of IP addresses for the interface.",
			Required:    true,
			// TODO Add field validation
		},
		"routes": {
			Optional:    true,
			Attributes:  tfsdk.ListNestedAttributes(RouteSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			Description: RouteSchema.Description,
		},
		"dhcp": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Indicates if DHCP should be used.",
		},
		"vlan_id": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "The VLAN’s ID. Must be a 16 bit unsigned integer.",
		},
		"mtu": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "The VLAN’s MTU. Must be a 32 bit unsigned integer.",
		},
		"vip": {
			Optional:    true,
			Attributes:  tfsdk.SingleNestedAttributes(VIPSchema.Attributes),
			Description: VIPSchema.Description,
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planVLAN VLAN) Data() (interface{}, error) {
	vlan := &v1alpha1.Vlan{}

	for _, vlanAddress := range planVLAN.Addresses {
		vlan.VlanAddresses = append(vlan.VlanAddresses, vlanAddress.Value)
	}
	for _, planVLANRoute := range planVLAN.Routes {
		route, err := planVLANRoute.Data()
		if err != nil {
			return &v1alpha1.Vlan{}, err
		}
		vlan.VlanRoutes = append(vlan.VlanRoutes, route.(*v1alpha1.Route))
	}
	if !planVLAN.DHCP.Null {
		vlan.VlanDHCP = planVLAN.DHCP.Value
	}
	if !planVLAN.VLANId.Null {
		vlan.VlanID = uint16(planVLAN.VLANId.Value)
	}
	if !planVLAN.MTU.Null {
		vlan.VlanMTU = uint32(planVLAN.MTU.Value)
	}
	if planVLAN.VIP != nil {
		vip, err := planVLAN.VIP.Data()
		if err != nil {
			return &v1alpha1.Vlan{}, err
		}
		vlan.VlanVIP = vip.(*v1alpha1.DeviceVIPConfig)
	}
	return vlan, nil
}

// VIP represent virtual shared IP configurations for network interfaces.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#devicevipconfig for more information.
type VIP struct {
	IP                   types.String `tfsdk:"ip"`
	EquinixMetalAPIToken types.String `tfsdk:"equinix_metal_api_token"`
	HetznerCloudAPIToken types.String `tfsdk:"hetzner_cloud_api_token"`
}

// VIPSchema represent virtual shared IP configurations for network interfaces.
var VIPSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Contains settings for configuring a Virtual Shared IP on an interface.",
	Attributes: map[string]tfsdk.Attribute{
		"ip": {
			Type:     types.StringType,
			Required: true,
			// TODO validate
			// ValidateFunc: validateIP,
			Description: "Specifies the IP address to be used.",
		},
		"equinix_metal_api_token": {
			Type:        types.StringType,
			Optional:    true,
			Description: "Specifies the Equinix Metal API Token.",
		},
		"hetzner_cloud_api_token": {
			Type:        types.StringType,
			Optional:    true,
			Description: "Specifies the Hetzner Cloud API Token.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planVIP VIP) Data() (interface{}, error) {
	vip := &v1alpha1.DeviceVIPConfig{
		SharedIP: planVIP.IP.Value,
	}
	if !planVIP.EquinixMetalAPIToken.Null {
		vip.EquinixMetalConfig = &v1alpha1.VIPEquinixMetalConfig{
			EquinixMetalAPIToken: planVIP.EquinixMetalAPIToken.Value,
		}
	}
	if !planVIP.HetznerCloudAPIToken.Null {
		vip.HCloudConfig = &v1alpha1.VIPHCloudConfig{
			HCloudAPIToken: planVIP.HetznerCloudAPIToken.Value,
		}
	}

	return vip, nil
}

// Route represents a network route.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#route for more information.
type Route struct {
	Network types.String `tfsdk:"network"`
	Gateway types.String `tfsdk:"gateway"`
	Source  types.String `tfsdk:"source"`
	Metric  types.String `tfsdk:"metric"`
}

// RouteSchema represents a network route.
var RouteSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents a list of routes.",
	Attributes: map[string]tfsdk.Attribute{
		"network": {
			Type:     types.StringType,
			Required: true,
			// TODO Validate
			// ValidateFunc: validateCIDR,
			Description: "The route’s network (destination).",
		},
		"gateway": {
			Type:     types.StringType,
			Optional: true,
			// TODO Validate
			// ValidateFunc: validateIP,
			Description: "The route’s gateway (if empty, creates link scope route).",
		},
		"source": {
			Type:     types.StringType,
			Optional: true,
			// TODO validate
			// ValidateFunc: validateIP,
			Description: "The route’s source address.",
		},
		"metric": {
			Type:        types.StringType,
			Optional:    true,
			Description: "The optional metric for the route.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planRoute Route) Data() (interface{}, error) {
	route := &v1alpha1.Route{
		RouteNetwork: planRoute.Network.Value,
	}

	if !planRoute.Gateway.Null {
		route.RouteGateway = planRoute.Gateway.Value
	}
	if !planRoute.Source.Null {
		route.RouteSource = planRoute.Source.Value
	}
	if !planRoute.Metric.Null {
		route.RouteSource = planRoute.Metric.Value
	}

	return route, nil
}

// Wireguard describes a network interface's Wireguard configuration and keys.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#devicewireguardconfig for more information.
type Wireguard struct {
	Peers      []WireguardPeer `tfsdk:"peer"`
	PublicKey  types.String    `tfsdk:"public_key"`
	PrivateKey types.String    `tfsdk:"private_key"`
}

// WireguardSchema describes a network interface's Wireguard configuration and keys.
var WireguardSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Contains settings for configuring Wireguard network interface.",
	Attributes: map[string]tfsdk.Attribute{
		"peer": {
			Required:    true,
			Attributes:  tfsdk.ListNestedAttributes(WireguardPeerSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			Description: WireguardPeerSchema.Description,
		},
		"public_key": {
			Type:        types.StringType,
			Computed:    true,
			Description: "Automatically derived from the private_key field.",
		},
		"private_key": {
			Type:      types.StringType,
			Sensitive: true,
			Optional:  true,
			Computed:  true,
			// TODO validate
			// ValidateFunc: validateKey,
			Description: "Specifies a private key configuration (base64 encoded). If one is not provided it is automatically generated and populated this field",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planWireguard Wireguard) Data() (interface{}, error) {
	wireguard := &v1alpha1.DeviceWireguardConfig{}

	for _, planPeer := range planWireguard.Peers {
		peer, err := planPeer.Data()
		if err != nil {
			return &v1alpha1.DeviceWireguardConfig{}, nil
		}
		wireguard.WireguardPeers = append(wireguard.WireguardPeers, peer.(*v1alpha1.DeviceWireguardPeer))
	}

	if !planWireguard.PrivateKey.Null {
		wireguard.WireguardPrivateKey = planWireguard.PrivateKey.Value
	}

	return wireguard, nil
}

// WireguardPeer describes an interface's Wireguard peers.
type WireguardPeer struct {
	AllowedIPs                  []types.String `tfsdk:"allowed_ips"`
	Endpoint                    types.String   `tfsdk:"endpoint"`
	PersistentKeepaliveInterval types.Int64    `tfsdk:"persistent_keepalive_interval"`
	PublicKey                   types.String   `tfsdk:"public_key"`
}

// WireguardPeerSchema describes an interface's Wireguard peers.
var WireguardPeerSchema tfsdk.Schema = tfsdk.Schema{
	Description: "A WireGuard device peer configuration.",
	Attributes: map[string]tfsdk.Attribute{
		"allowed_ips": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required:    true,
			Description: "AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.",
			// TODO add validator
			// ValidateFunc: validateCIDR,
		},
		"endpoint": {
			Type:     types.StringType,
			Required: true,
			// TODO Add validator
			//ValidateFunc: validateEndpoint64Type,
			Description: "Specifies the endpoint of this peer entry.",
		},
		"persistent_keepalive_interval": {
			Type:     types.Int64Type,
			Optional: true,
			// TODO Add validator, assert it is positive and within the expected range
			/*
				ValidateFunc: func(value interface{}, key string) (warns []stringType, errs []error) {
					v := value.(int)
					if v < 0 {
						errs = append(errs, fmt.Errorf("%s: Persistent keepalive interval must be a positive integer, got %d", key, v))
					}
					return
				},
			*/
			Description: "Specifies the persistent keepalive interval for this peer. Provided in seconds.",
		},
		"public_key": {
			Type:     types.StringType,
			Required: true,
			// TODO: Add validator for ValidateFunc: validateKey,
			Description: "Specifies the public key of this peer.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planPeer WireguardPeer) Data() (interface{}, error) {
	peer := &v1alpha1.DeviceWireguardPeer{
		WireguardPublicKey: planPeer.PublicKey.Value,
		WireguardEndpoint:  planPeer.Endpoint.Value,
	}

	for _, ip := range planPeer.AllowedIPs {
		peer.WireguardAllowedIPs = append(peer.WireguardAllowedIPs, ip.Value)
	}

	if !planPeer.PersistentKeepaliveInterval.Null {
		peer.WireguardPersistentKeepaliveInterval = time.Duration(planPeer.PersistentKeepaliveInterval.Value) * time.Second
	}

	return peer, nil
}

// MachineControlPlane configures options pertaining to the Kubernetes control plane that's installed onto the machine.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#machinecontrolplaneconfig for more information.
type MachineControlPlane struct {
	ControllerManagerDisabled types.Bool `tfsdk:"controller_manager_disabled"`
	SchedulerDisabled         types.Bool `tfsdk:"scheduler_disabled"`
}

// MachineControlPlaneSchema configures options pertaining to the Kubernetes control plane that's installed onto the machine.
var MachineControlPlaneSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures options pertaining to the Kubernetes control plane that's installed onto the machine",
	Attributes: map[string]tfsdk.Attribute{
		"controller_manager_disabled": {
			Type:     types.BoolType,
			Optional: true,
			Description: "Disable kube-controller-manager on the node.	",
		},
		"scheduler_disabled": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Disable kube-scheduler on the node.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planMachineControlPlane MachineControlPlane) Data() (interface{}, error) {
	controlConf := &v1alpha1.MachineControlPlaneConfig{
		MachineControllerManager: &v1alpha1.MachineControllerManagerConfig{},
		MachineScheduler:         &v1alpha1.MachineSchedulerConfig{},
	}

	if !planMachineControlPlane.ControllerManagerDisabled.Null {
		controlConf.MachineControllerManager.MachineControllerManagerDisabled = planMachineControlPlane.ControllerManagerDisabled.Value
	}

	if !planMachineControlPlane.ControllerManagerDisabled.Null {
		controlConf.MachineScheduler.MachineSchedulerDisabled = planMachineControlPlane.SchedulerDisabled.Value
	}

	return controlConf, nil
}

// PartitionData represents the options for a disk partition.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#diskpartition for more information.
type PartitionData struct {
	Size       types.String `tfsdk:"size"`
	MountPoint types.String `tfsdk:"mount_point"`
}

// PartitionSchema represents the options for a disk partition.
var PartitionSchema = tfsdk.Schema{
	MarkdownDescription: `Represents the options for a disk partition.`,
	Attributes: map[string]tfsdk.Attribute{
		"size": {
			Required: true,
			MarkdownDescription: `The size of partition: either bytes or human readable representation.
If ` + "`size:`" + `is omitted, the partition is sized to occupy the full disk.`,
			Type: types.StringType,
		},
		"mount_point": {
			Required:    true,
			Description: "Where the partition will be mounted.",
			Type:        types.StringType,
		},
	},
}

func (partition PartitionData) Data() (any, error) {
	size, err := humanize.ParseBytes(partition.Size.Value)
	if err != nil {
		return nil, err
	}

	part := &v1alpha1.DiskPartition{
		DiskSize:       v1alpha1.DiskSize(size),
		DiskMountPoint: partition.Size.Value,
	}

	return part, nil
}

func (partition *PartitionData) Read(part any) error {
	diskPartition := part.(*v1alpha1.DiskPartition)

	size, err := diskPartition.DiskSize.MarshalYAML()
	if err != nil {
		return err
	}
	partition.Size = types.String{Value: size.(string)}

	return nil
}

// MachineDiskData represents the options available for partitioning, formatting, and mounting extra disks.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#machinedisk for more information.
type MachineDiskData struct {
	DeviceName types.String    `tfsdk:"device_name"`
	Partitions []PartitionData `tfsdk:"partitions"`
}

// MachineDiskSchema represents the options available for partitioning, formatting, and mounting extra disks.
var MachineDiskSchema = tfsdk.Schema{
	MarkdownDescription: "Represents partitioning for disks on the machine.",
	Attributes: map[string]tfsdk.Attribute{
		"device_name": {
			Required:    true,
			Description: "Block device name.",
			Type:        types.StringType,
		},
		"partitions": {
			Required:    true,
			Description: PartitionSchema.MarkdownDescription,
			Attributes:  tfsdk.ListNestedAttributes(PartitionSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
		},
	},
}

func (diskData MachineDiskData) Data() (any, error) {
	disk := v1alpha1.MachineDisk{
		DeviceName: diskData.DeviceName.Value,
	}

	for _, partition := range diskData.Partitions {
		part, err := partition.Data()
		if err != nil {
			return nil, err
		}
		disk.DiskPartitions = append(disk.DiskPartitions, part.(*v1alpha1.DiskPartition))
	}

	return disk, nil
}

func (data *MachineDiskData) Read(diskData any) error {
	machineDisk := diskData.(v1alpha1.MachineDisk)

	data.DeviceName.Value = machineDisk.DeviceName
	for _, partition := range machineDisk.DiskPartitions {
		part := PartitionData{}
		err := part.Read(partition)
		if err != nil {
			return err
		}
		data.Partitions = append(data.Partitions, part)
	}

	return nil
}

// EncryptionData specifies system disk partitions encryption settings.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#systemdiskencryptionconfig for more information.
type EncryptionData struct {
	State     *EncryptionConfigData `tfsdk:"state"`
	Ephemeral *EncryptionConfigData `tfsdk:"ephemeral"`
}

// EncryptionSchema specifies system disk partitions encryption settings.
var EncryptionSchema = tfsdk.Schema{
	MarkdownDescription: "Specifies system disk partition encryption settings.",
	Attributes: map[string]tfsdk.Attribute{
		"state": {
			Optional:    true,
			Description: EncryptionConfigSchema.MarkdownDescription,
			Attributes:  tfsdk.SingleNestedAttributes(EncryptionConfigSchema.Attributes),
		},
		"ephemeral": {
			Optional:    true,
			Description: EncryptionConfigSchema.MarkdownDescription,
			Attributes:  tfsdk.SingleNestedAttributes(EncryptionConfigSchema.Attributes),
		},
		// TODO requires at least one of
	},
}

func (encryptionData EncryptionData) Data() (any, error) {
	encryption := &v1alpha1.SystemDiskEncryptionConfig{}

	if encryptionData.State != nil {
		state, err := encryptionData.State.Data()
		if err != nil {
			return nil, err
		}
		encryption.StatePartition = state.(*v1alpha1.EncryptionConfig)
	}

	if encryptionData.Ephemeral != nil {
		ephemeral, err := encryptionData.Ephemeral.Data()
		if err != nil {
			return nil, err
		}
		encryption.EphemeralPartition = ephemeral.(*v1alpha1.EncryptionConfig)
	}

	return encryption, nil
}

func (data *EncryptionData) Read(diskData any) error {
	encryptionConfig := diskData.(*v1alpha1.SystemDiskEncryptionConfig)

	if encryptionConfig.StatePartition != nil {
		data.State = &EncryptionConfigData{}
		err := data.State.Read(encryptionConfig.StatePartition)
		if err != nil {
			return err
		}
	}

	if encryptionConfig.EphemeralPartition != nil {
		ephemeralData := &EncryptionConfigData{}
		err := ephemeralData.Read(encryptionConfig.EphemeralPartition)
		if err != nil {
			return err
		}
		data.Ephemeral = ephemeralData
	}

	return nil
}

// EncryptionConfigData represents partition encryption settings.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#encryptionconfig for more information.
type EncryptionConfigData struct {
	Provider    types.String   `tfsdk:"provider"`
	Keys        []KeyConfig    `tfsdk:"keys"`
	Cipher      types.String   `tfsdk:"cipher"`
	KeySize     types.Int64    `tfsdk:"keysize"`
	BlockSize   types.Int64    `tfsdk:"blocksize"`
	PerfOptions []types.String `tfsdk:"perf_options"`
}

// EncryptionConfigSchema represents partition encryption settings.
var EncryptionConfigSchema = tfsdk.Schema{
	MarkdownDescription: "Represents partition encryption settings.",
	Attributes: map[string]tfsdk.Attribute{
		"provider": {
			Required:    true,
			Description: "Encryption provider to use for the encryption.",
			Type:        types.StringType,
		},
		"keys": {
			Required:    true,
			Description: KeySchema.MarkdownDescription,
			Attributes:  tfsdk.ListNestedAttributes(KeySchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
		},
		"cipher": {
			Optional:    true,
			Description: "Cipher kind to use for the encryption. Depends on the encryption provider.",
			Type:        types.StringType,
		},
		"keysize": {
			Optional:    true,
			Description: "Defines the encryption key size.",
			Type:        types.Int64Type,
		},
		"blocksize": {
			Optional:    true,
			Description: "Defines the encryption block size.",
			Type:        types.Int64Type,
		},
		"perf_options": {
			Optional:    true,
			Description: "Additional --perf parameters for LUKS2 encryption.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
		},
	},
}

func (encryptionData EncryptionConfigData) Data() (any, error) {
	encryptionConfig := &v1alpha1.EncryptionConfig{
		EncryptionProvider: encryptionData.Provider.Value,
	}

	for _, key := range encryptionData.Keys {
		k, err := key.Data()
		if err != nil {
			return nil, err
		}
		encryptionConfig.EncryptionKeys = append(encryptionConfig.EncryptionKeys, k.(*v1alpha1.EncryptionKey))
	}

	if !encryptionData.Cipher.Null {
		encryptionConfig.EncryptionCipher = encryptionData.Cipher.Value
	}

	if !encryptionData.KeySize.Null {
		encryptionConfig.EncryptionKeySize = uint(encryptionData.KeySize.Value)
	}

	if !encryptionData.BlockSize.Null {
		encryptionConfig.EncryptionBlockSize = uint64(encryptionData.BlockSize.Value)
	}

	for _, opt := range encryptionData.PerfOptions {
		encryptionConfig.EncryptionPerfOptions = append(encryptionConfig.EncryptionPerfOptions, opt.Value)
	}

	return encryptionConfig, nil
}

func (data *EncryptionConfigData) Read(encryptionData any) error {
	partEncryptionConfig := encryptionData.(*v1alpha1.EncryptionConfig)

	data.Provider.Value = partEncryptionConfig.EncryptionProvider

	for _, key := range partEncryptionConfig.EncryptionKeys {
		keyconfig := KeyConfig{}
		err := keyconfig.Read(key)
		if err != nil {
			return err
		}
		data.Keys = append(data.Keys, keyconfig)
	}

	if partEncryptionConfig.EncryptionCipher != *new(string) {
		data.Cipher.Value = partEncryptionConfig.EncryptionCipher
	}

	if partEncryptionConfig.EncryptionKeySize != *new(uint) {
		data.KeySize.Value = int64(partEncryptionConfig.EncryptionKeySize)
	}

	if partEncryptionConfig.EncryptionBlockSize != *new(uint64) {
		data.BlockSize.Value = int64(partEncryptionConfig.EncryptionBlockSize)
	}

	for _, opt := range partEncryptionConfig.EncryptionPerfOptions {
		data.PerfOptions = append(data.PerfOptions, types.String{Value: opt})
	}

	return nil
}

// KeyConfig represents configuration for disk encryption key.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#encryptionkey for more information.
type KeyConfig struct {
	KeyStatic types.String `tfsdk:"key_static"`
	NodeID    types.Bool   `tfsdk:"node_id"`
	Slot      types.Int64  `tfsdk:"slot"`
}

// KeySchema represents configuration for disk encryption key.
var KeySchema = tfsdk.Schema{
	MarkdownDescription: "Specifies system disk partition encryption settings.",
	Attributes: map[string]tfsdk.Attribute{
		// TODO have key_static and node_id mutually exclusive
		"key_static": {
			Optional:    true,
			Description: "Represents a throw away key type.",
			Type:        types.StringType,
		},
		"node_id": {
			Optional:    true,
			Description: "Represents a deterministically generated key from the node UUID and PartitionLabel. Setting this value to true will enable it.",
			Type:        types.BoolType,
		},
		"slot": {
			Required:    true,
			Description: "Defines the encryption block size.",
			Type:        types.Int64Type,
		},
	},
}

func (keyData KeyConfig) Data() (any, error) {
	encryptionKey := &v1alpha1.EncryptionKey{
		KeySlot: int(keyData.Slot.Value),
	}

	if !keyData.KeyStatic.Null {
		encryptionKey.KeyStatic = &v1alpha1.EncryptionKeyStatic{
			KeyData: keyData.KeyStatic.Value,
		}
	}

	if !keyData.NodeID.Null {
		encryptionKey.KeyNodeID = &v1alpha1.EncryptionKeyNodeID{}
	}

	return encryptionKey, nil
}

func (data *KeyConfig) Read(keyData any) error {
	key := keyData.(*v1alpha1.EncryptionKey)

	data.Slot.Value = int64(key.KeySlot)

	if key.KeySlot != *new(int) {
		data.Slot.Value = int64(key.KeySlot)
	}

	if key.KeyNodeID != nil {
		data.NodeID.Value = true
	}

	if key.KeyStatic != nil {
		data.KeyStatic.Value = key.KeyStatic.KeyData
	}

	return nil
}

// AdmissionPluginConfig configures pod admssion rules on the kubelet64Type, denying execution to pods that don't fit them.
type AdmissionPluginConfig struct {
	Name          types.String `tfsdk:"name"`
	Configuration types.String `tfsdk:"configuration"`
}

// AdmissionPluginSchema configures pod admssion rules on the kubelet64Type, denying execution to pods that don't fit them.
var AdmissionPluginSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures pod admssion rules on the kubelet64Type, denying execution to pods that don't fit them.",
	Attributes: map[string]tfsdk.Attribute{
		"name": {
			Type:        types.StringType,
			Required:    true,
			Description: "Name is the name of the admission controller. It must match the registered admission plugin name.",
			// TODO Validate it is a properly formed name
		},
		"configuration": {
			Type:        types.StringType,
			Required:    true,
			Description: "Configuration is an embedded configuration object to be used as the plugin’s configuration.",
			// TODO Validate it is a properly formed YAML
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planAdmissionPluginConfig AdmissionPluginConfig) Data() (interface{}, error) {
	var admissionConfig v1alpha1.Unstructured

	if err := yaml.Unmarshal([]byte(planAdmissionPluginConfig.Configuration.Value), &admissionConfig); err != nil {
		return &v1alpha1.AdmissionPluginConfig{}, nil
	}

	admissionPluginConfig := &v1alpha1.AdmissionPluginConfig{
		PluginName:          planAdmissionPluginConfig.Name.Value,
		PluginConfiguration: admissionConfig,
	}
	return admissionPluginConfig, nil
}

// ProxyConfig configures the Kubernetes control plane's kube-proxy.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#proxyconfig for more information.
type ProxyConfig struct {
	Image types.String `tfsdk:"image"`

	Mode      types.String            `tfsdk:"mode"`
	Disabled  types.Bool              `tfsdk:"is_disabled"`
	ExtraArgs map[string]types.String `tfsdk:"extra_args"`
}

// ProxyConfigSchema configures the Kubernetes control plane's kube-proxy.
var ProxyConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kube proxy configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:        types.StringType,
			Required:    true,
			Description: "The container image used in the kube-proxy manifest.",
			// TODO validate
			// ValidateFunc: validateImage,
		},
		"mode": {
			Type:        types.StringType,
			Optional:    true,
			Description: "The container image used in the kube-proxy manifest.",
			// TODO Validate it's a valid mode
		},
		"is_disabled": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Disable kube-proxy deployment on cluster bootstrap.",
		},
		"extra_args": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Extra arguments to supply to kube-proxy.",
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planProxy ProxyConfig) Data() (interface{}, error) {
	proxy := &v1alpha1.ProxyConfig{}
	if !planProxy.Image.Null {
		proxy.ContainerImage = planProxy.Image.Value
	}
	if !planProxy.Disabled.Null {
		proxy.Disabled = planProxy.Disabled.Value
	}
	if !planProxy.Mode.Null {
		proxy.ModeConfig = planProxy.Mode.Value
	}
	proxy.ExtraArgsConfig = map[string]string{}
	for arg, value := range planProxy.ExtraArgs {
		proxy.ExtraArgsConfig[arg] = value.Value
	}

	return proxy, nil
}

// ControlPlaneConfig provides options for configuring the Kubernetes control plane.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#controlplaneconfig for more information.
type ControlPlaneConfig struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	LocalAPIServerPort types.Int64  `tfsdk:"local_api_server_port"`
}

// ControlPlaneConfigSchema provides options for configuring the Kubernetes control plane.
var ControlPlaneConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the control plane configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"endpoint": {
			Type:        types.StringType,
			Optional:    true,
			Description: "Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.",
			// TODO Verify well formed endpoint
		},
		"local_api_server_port": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "The port that the API server listens on internally. This may be different than the port portion listed in the endpoint field.",
			// TODO Verify in correct port range
		},
	},
}

// APIServerConfig configures the Kubernetes control plane's apiserver.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#apiserverconfig for more information.
type APIServerConfig struct {
	Image            types.String            `tfsdk:"image"`
	ExtraArgs        map[string]types.String `tfsdk:"extra_args"`
	ExtraVolumes     []VolumeMount           `tfsdk:"extra_volumes"`
	Env              map[string]types.String `tfsdk:"env"`
	CertSANS         []types.String          `tfsdk:"cert_sans"`
	DisablePSP       types.Bool              `tfsdk:"disable_pod_security_policy"`
	AdmissionPlugins []AdmissionPluginConfig `tfsdk:"admission_control"`
}

// APIServerConfigSchema configures the Kubernetes control plane's apiserver.
var APIServerConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kube apiserver configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:        types.StringType,
			Required:    true,
			Description: "The container image used in the API server manifest.",
			// TODO validation
			// ValidateFunc: validateImage,
		},
		"extra_args": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Extra arguments to supply to the API server.",
		},

		"extra_volumes": {
			Optional:    true,
			Description: VolumeMountSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(VolumeMountSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
		},
		"env": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "The env field allows for the addition of environment variables for the control plane component.",
		},
		// TODO validate IPs
		"cert_sans": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required:    true,
			Description: "Extra certificate subject alternative names for the API server’s certificate.",
		},
		"disable_pod_security_policy": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Disable PodSecurityPolicy in the API server and default manifests.",
		},

		"admission_control": {
			Optional:    true,
			Description: AdmissionPluginSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(AdmissionPluginSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planAPIServer APIServerConfig) Data() (interface{}, error) {
	apiServer := &v1alpha1.APIServerConfig{}

	if !planAPIServer.Image.Null {
		apiServer.ContainerImage = planAPIServer.Image.Value
	}
	apiServer.ExtraArgsConfig = map[string]string{}
	for arg, value := range planAPIServer.ExtraArgs {
		apiServer.ExtraArgsConfig[arg] = value.Value
	}
	if !planAPIServer.DisablePSP.Null {
		apiServer.DisablePodSecurityPolicyConfig = planAPIServer.DisablePSP.Value
	}

	for i, pluginYaml := range planAPIServer.AdmissionPlugins {
		apiServer.AdmissionControlConfig = append(apiServer.AdmissionControlConfig, &v1alpha1.AdmissionPluginConfig{
			PluginName: pluginYaml.Name.Value,
		})

		var plugin v1alpha1.Unstructured
		if err := yaml.Unmarshal([]byte(pluginYaml.Configuration.Value), &plugin); err != nil {
			return &v1alpha1.APIServerConfig{}, err
		}
		apiServer.AdmissionControlConfig[i].PluginConfiguration = plugin
	}

	for _, san := range planAPIServer.CertSANS {
		apiServer.CertSANs = append(apiServer.CertSANs, san.Value)
	}
	apiServer.EnvConfig = map[string]string{}
	for arg, value := range planAPIServer.Env {
		apiServer.EnvConfig[arg] = value.Value
	}
	for _, vol := range planAPIServer.ExtraVolumes {
		d, err := vol.Data()
		if err != nil {
			return &v1alpha1.APIServerConfig{}, err
		}
		apiServer.ExtraVolumesConfig = append(apiServer.ExtraVolumesConfig, d.(v1alpha1.VolumeMountConfig))
	}

	return apiServer, nil
}

// File describes a machine file and it's contents to be written onto the node's filesystem.
type File struct {
	Content     types.String `tfsdk:"content"`
	Permissions types.Int64  `tfsdk:"permissions"`
	Path        types.String `tfsdk:"path"`
	Op          types.String `tfsdk:"op"`
}

// FileSchema describes a machine file and it's contents to be written onto the node's filesystem.
var FileSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Describes a machine's files and it's contents and how it will be written to the node's filesystem.",
	Attributes: map[string]tfsdk.Attribute{
		"content": {
			Type:        types.StringType,
			Required:    true,
			Description: "The file's content. Not required to be base64 encoded.",
		},
		"permissions": {
			Type:        types.Int64Type,
			Required:    true,
			Description: "Unix permission for the file",
			// TODO validate
			/*
				ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
					v := value.(int)
					if v < 0 {
						errs = append(errs, fmt.Errorf("Persistent keepalive interval must be a positive integer, got %d", v))
					}
					return
				},
			*/
		},
		"path": {
			Type:        types.StringType,
			Required:    true,
			Description: "Full path for the file to be created at.",
			// TODO: Add validation for path correctness
		},
		"op": {
			Type:        types.StringType,
			Required:    true,
			Description: "Mode for the file. Can be one of create, append and overwrite.",
			// TODO validate
			/*
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
			*/
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planFile File) Data() (interface{}, error) {
	return &v1alpha1.MachineFile{
		FileContent:     planFile.Content.Value,
		FilePermissions: v1alpha1.FileMode(planFile.Permissions.Value),
		FilePath:        planFile.Path.Value,
		FileOp:          planFile.Op.Value,
	}, nil
}

// InlineManifest describes inline bootstrap manifests for the user.
type InlineManifest struct {
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
}

// InlineManifestSchema describes inline bootstrap manifests for the user.
var InlineManifestSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Describes inline bootstrap manifests for the user. These will get automatically deployed as part of the bootstrap.",
	Attributes: map[string]tfsdk.Attribute{
		"name": {
			Type:        types.StringType,
			Required:    true,
			Description: "The manifest's name.",
		},
		"content": {
			Type:        types.StringType,
			Required:    true,
			Description: "The manifest's content. Must be a valid kubernetes YAML.",
			// TODO validate InlineManifestSchema content field
		},
	},
}

// Data copies data from terraform state types to talos types.
func (planManifest InlineManifest) Data() (interface{}, error) {
	manifest := v1alpha1.ClusterInlineManifest{}

	if planManifest.Name.Value != "" {
		manifest.InlineManifestName = planManifest.Name.Value
		manifest.InlineManifestContents = planManifest.Content.Value
	}

	return manifest, nil
}

// Read copies data from talos types to terraform state types.
func (planManifest *InlineManifest) Read(talosInlineManifest interface{}) error {
	manifest := talosInlineManifest.(v1alpha1.ClusterInlineManifest)
	if manifest.InlineManifestName != "" {
		planManifest.Name = types.String{Value: manifest.InlineManifestName}
		planManifest.Content = types.String{Value: manifest.InlineManifestContents}
	}

	return nil
}

type nodeResourceData interface {
	TalosData(*v1alpha1.Config) (*v1alpha1.Config, error)
	ReadInto(*v1alpha1.Config) error
	Generate() error
}

type readData struct {
	ConfigIP   string
	BaseConfig string
}

func readConfig[N nodeResourceData](ctx context.Context, nodeData N, data readData) (out *v1alpha1.Config, errDesc string, err error) {
	host := net.JoinHostPort(data.ConfigIP, strconv.Itoa(talosPort))

	input := generate.Input{}
	if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
		return nil, "Unable to marshal node's base_config data into it's generate.Input struct.", err
	}

	conn, err := secureConn(ctx, input, host)
	if err != nil {
		return nil, "Unable to make a secure connection to read the node's Talos config.", err
	}

	defer conn.Close()
	client := resource.NewResourceServiceClient(conn)
	resourceResp, err := client.Get(ctx, &resource.GetRequest{
		Type:      "MachineConfig",
		Namespace: "config",
		Id:        "v1alpha1",
	})
	if err != nil {
		return nil, "Error getting Machine Configuration", err
	}

	if len(resourceResp.Messages) < 1 {
		return nil, "Invalid message count.",
			fmt.Errorf("invalid message count from the Talos resource get request. Expected > 1 but got %d", len(resourceResp.Messages))
	}

	out = &v1alpha1.Config{}
	err = yaml.Unmarshal(resourceResp.Messages[0].Resource.Spec.Yaml, out)
	if err != nil {
		return nil, "Unable to unmarshal Talos configuration into it's struct.", err
	}

	return
}

type configData struct {
	Bootstrap   bool
	CreateNode  bool
	Mode        machine.ApplyConfigurationRequest_Mode
	ConfigIP    string
	BaseConfig  string
	MachineType machinetype.Type
	Network     string
	MAC         string
}

func genConfig[N nodeResourceData](machineType machinetype.Type, input *generate.Input, nodeData *N) (out string, err error) {
	cfg, err := generate.Config(machineType, input)
	if err != nil {
		err = fmt.Errorf("failed to generate Talos configuration struct for node: %w", err)
		return
	}

	(*nodeData).Generate()
	newCfg, err := (*nodeData).TalosData(cfg)
	if err != nil {
		err = fmt.Errorf("failed to generate configuration: %w", err)
		return
	}

	var confYaml []byte
	confYaml, err = newCfg.Bytes()
	if err != nil {
		err = fmt.Errorf("failed to generate config yaml: %w", err)
		return
	}

	out = string(regexp.MustCompile(`\s*#.*`).ReplaceAll(confYaml, nil))
	return
}

func applyConfig[N nodeResourceData](ctx context.Context, nodeData *N, data configData) (out string, errDesc string, err error) {
	input := generate.Input{}
	if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
		return "", "Failed to unmarshal input bundle", err
	}

	yaml, err := genConfig(data.MachineType, &input, nodeData)
	if err != nil {
		return "", "error rendering configuration YAML", err
	}

	var conn *grpc.ClientConn
	if data.CreateNode {
		network := data.Network
		mac := data.MAC

		dhcpIP, err := lookupIP(ctx, network, mac)
		if err != nil {
			return "", "Error looking up node IP", err
		}

		host := net.JoinHostPort(dhcpIP.String(), strconv.Itoa(talosPort))
		conn, err = insecureConn(ctx, host)
		if err != nil {
			return "", "Unable to make insecure connection to Talos machine. Ensure it is in maintainence mode.", err
		}
	} else {
		ip := data.ConfigIP
		host := net.JoinHostPort(ip, strconv.Itoa(talosPort))
		input := generate.Input{}
		if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
			return "", "Unable to unmarshal BaseConfig json into a Talos Input struct.", err
		}

		conn, err = secureConn(ctx, input, host)
		if err != nil {
			return "", "Unable to make secure connection to the Talos machine.", err
		}
	}

	defer conn.Close()
	client := machine.NewMachineServiceClient(conn)
	_, err = client.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: []byte(yaml),
		Mode: machine.ApplyConfigurationRequest_Mode(data.Mode),
	})
	if err != nil {
		return "", "Error applying configuration", err
	}

	if data.MachineType == machinetype.TypeControlPlane && data.Bootstrap {
		// Wait for time to be synchronised.
		time.Sleep(5 * time.Second)

		ip := data.ConfigIP
		host := net.JoinHostPort(ip, strconv.Itoa(talosPort))
		input := generate.Input{}
		if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
			return "", "Unable to unmarshal BaseConfig json into a Talos Input struct.", err
		}

		conn, err := secureConn(ctx, input, host)
		if err != nil {
			return "", "Unable to make secure connection to the Talos machine.", err
		}
		defer conn.Close()
		client := machine.NewMachineServiceClient(conn)
		_, err = client.Bootstrap(ctx, &machine.BootstrapRequest{})
		if err != nil {
			return "", "Error attempting to bootstrap the machine.", err
		}
	}

	return
}

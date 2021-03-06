package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var TalosConfigSchema = tfsdk.Schema{
	MarkdownDescription: "",
	Attributes: map[string]tfsdk.Attribute{
		"install": {
			Required:    true,
			Description: InstallSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(InstallSchema.Attributes),
		},
		"cert_sans": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
			// TODO validation
			Description: "Extra certificate subject alternative names for the machine’s certificate.",
		},
		"control_plane": {
			Optional:    true,
			Description: ControlPlaneConfigSchema.Description,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Attributes: tfsdk.SingleNestedAttributes(ControlPlaneConfigSchema.Attributes),
		},
		"kubelet": {
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Description: KubeletConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(KubeletConfigSchema.Attributes),
		},
		"pods": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
			// TODO validation
			Description: "Used to provide static pod definitions to be run by the kubelet directly bypassing the kube-apiserver.",
		},
		"network": {
			Required:    true,
			Description: NetworkConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(NetworkConfigSchema.Attributes),
		},
		"disks": {
			Optional:    true,
			Description: MachineDiskSchema.MarkdownDescription,
			Attributes:  tfsdk.ListNestedAttributes(MachineDiskSchema.Attributes),
		},
		"files": {
			Optional:    true,
			Description: FileSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(FileSchema.Attributes),
		},
		"env": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Allows for the addition of environment variables. All environment variables are set on PID 1 in addition to every service.",
		},
		"time": {
			Description: TimeConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(TimeConfigSchema.Attributes),
			Optional:    true,
		},
		"sysctls": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Used to configure the machine’s sysctls.",
		},
		"sysfs": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Used to configure the machine’s sysctls.",
		},
		"registry": {
			Optional:    true,
			Description: RegistrySchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(RegistrySchema.Attributes),
		},
		"encryption": {
			Optional:    true,
			Description: EncryptionSchema.MarkdownDescription,
			Attributes:  tfsdk.SingleNestedAttributes(EncryptionSchema.Attributes),
		},
		// features not implemented
		"udev": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "Configures the udev system.",
			Optional:    true,
		},
		"logging": {
			Description: LoggingConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(LoggingConfigSchema.Attributes),
			Optional:    true,
		},
		"kernel": {
			Description: KernelConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(KernelConfigSchema.Attributes),
			Optional:    true,
		},

		"control_plane_config": {
			Optional:    true,
			Description: MachineControlPlaneSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(MachineControlPlaneSchema.Attributes),
		},

		// clustername already filled
		// cluster_network not implemented
		"apiserver": {
			Optional:    true,
			Description: APIServerConfigSchema.Description,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Attributes: tfsdk.SingleNestedAttributes(APIServerConfigSchema.Attributes),
		},
		"controller_manager": {
			Description: ControllerManagerConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(ControllerManagerConfigSchema.Attributes),
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Optional: true,
		},
		"proxy": {
			Optional:    true,
			Description: ProxyConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(ProxyConfigSchema.Attributes),
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"scheduler": {
			Description: SchedulerConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(SchedulerConfigSchema.Attributes),
			Optional:    true,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"discovery": {
			Description: ClusterDiscoveryConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(ClusterDiscoveryConfigSchema.Attributes),
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Optional: true,
		},
		"etcd": {
			Description: EtcdConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(EtcdConfigSchema.Attributes),
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Optional: true,
		},
		"coredns": {
			Description: CoreDNSConfigSchema.MarkdownDescription,
			Attributes:  tfsdk.SingleNestedAttributes(CoreDNSConfigSchema.Attributes),
			Optional:    true,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"external_cloud_provider": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "Contains external cloud provider configuration.",
			Optional:    true,
		},
		"extra_manifest_headers": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Description: "A map of key value pairs that will be added while fetching the extraManifests. ",
			Optional:    true,
		},
		"extra_manifests": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "A list of urls that point to additional manifests. These will get automatically deployed as part of the bootstrap.",
			Optional:    true,
		},
		// TODO Add verification function confirming it's a correct manifest that can be downloaded.
		"inline_manifests": {
			Optional:    true,
			Description: InlineManifestSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(InlineManifestSchema.Attributes),
		},
		"admin_kube_config": {
			Description: AdminKubeconfigConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(AdminKubeconfigConfigSchema.Attributes),
			Optional:    true,
		},
		"allow_scheduling_on_masters": {
			Type:     types.BoolType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Description: "Allows running workload on master nodes.",
		},
	},
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

// KubeletConfigSchema represents the kubelet's config values.
var KubeletConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kubelet's config values.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
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
			Attributes:  tfsdk.ListNestedAttributes(ExtraMountSchema.Attributes),
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

// CNISchema represents CNI info.
var CNISchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents CNI options.",
	Attributes: map[string]tfsdk.Attribute{
		"name": {
			Required: true,
			Type:     types.StringType,
		},
		"urls": {
			Required: true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
		},
	},
}

// NetworkKubeSpanSchema describes KubeSpan configuration.
var NetworkKubeSpanSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Describes Talos KubeSpan configuration.",
	Attributes: map[string]tfsdk.Attribute{
		"enabled": {
			Required:    true,
			Description: "Enable the KubeSpan feature.",
			Type:        types.BoolType,
		},
		"allow_peer_down_bypass": {
			Optional:    true,
			Type:        types.BoolType,
			Description: "Skip sending traffic via KubeSpan if the peer connection state is not up.",
		},
	},
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
			Attributes:  tfsdk.MapNestedAttributes(RegistryConfigSchema.Attributes),
		},
	},
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

var NetworkConfigSchema = tfsdk.Schema{
	MarkdownDescription: "Represents node network configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"hostname": {
			Type:        types.StringType,
			Optional:    true,
			Description: "Used to statically set the hostname for the machine.",
		},
		"devices": {
			Optional:    true,
			Description: NetworkDeviceSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(NetworkDeviceSchema.Attributes),
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
			Optional:    true,
			Description: NetworkKubeSpanSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(NetworkKubeSpanSchema.Attributes),
		},
	},
}

// NetworkDeviceSchema describes a Talos Device configuration.
var NetworkDeviceSchema tfsdk.Schema = tfsdk.Schema{
	Description:         "Describes a Talos network device configuration. The map's key is the interface name.",
	MarkdownDescription: "",
	Attributes: map[string]tfsdk.Attribute{
		"name": {
			Type:        types.StringType,
			Required:    true,
			Description: "Network device's Linux interface name.",
		},
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
			Attributes:  tfsdk.ListNestedAttributes(RouteSchema.Attributes),
			Description: RouteSchema.Description,
		},
		"bond": {
			Optional:    true,
			Attributes:  tfsdk.SingleNestedAttributes(BondSchema.Attributes),
			Description: BondSchema.Description,
		},
		"vlans": {
			Optional:    true,
			Attributes:  tfsdk.ListNestedAttributes(VLANSchema.Attributes),
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
			Attributes:  tfsdk.SingleNestedAttributes(DHCPOptionsSchema.Attributes),
			Description: DHCPOptionsSchema.Description,
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

// BondSchema contains the various options for configuring a bonded interface.
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
			Attributes:  tfsdk.ListNestedAttributes(RouteSchema.Attributes),
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

// RouteSchema represents a network route.
var RouteSchema tfsdk.Schema = tfsdk.Schema{
	Description:         "Represents a list of routes.",
	MarkdownDescription: "",
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
			Type:        types.Int64Type,
			Optional:    true,
			Description: "The optional metric for the route.",
		},
	},
}

// WireguardSchema describes a network interface's Wireguard configuration and keys.
var WireguardSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Contains settings for configuring Wireguard network interface.",
	Attributes: map[string]tfsdk.Attribute{
		"peers": {
			Required:    true,
			Attributes:  tfsdk.ListNestedAttributes(WireguardPeerSchema.Attributes),
			Description: WireguardPeerSchema.Description,
		},
		"firewall_mark": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "Firewall mark for wireguard packets.",
		},
		"listen_port": {
			Type:        types.Int64Type,
			Optional:    true,
			Description: "Listening port for if this node should be a wireguard server.",
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
			Attributes:  tfsdk.ListNestedAttributes(PartitionSchema.Attributes),
		},
	},
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

// EncryptionConfigSchema represents partition encryption settings.
var EncryptionConfigSchema = tfsdk.Schema{
	MarkdownDescription: "Represents partition encryption settings.",
	Attributes: map[string]tfsdk.Attribute{
		"crypt_provider": {
			Required:    true,
			Description: "Encryption provider to use for the encryption.",
			Type:        types.StringType,
		},
		"keys": {
			Required:    true,
			Description: KeySchema.MarkdownDescription,
			Attributes:  tfsdk.ListNestedAttributes(KeySchema.Attributes),
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

// APIServerConfigSchema configures the Kubernetes control plane's apiserver.
var APIServerConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kube apiserver configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
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
			Attributes:  tfsdk.ListNestedAttributes(VolumeMountSchema.Attributes),
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
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Description: "Extra certificate subject alternative names for the API server’s certificate.",
		},
		"disable_pod_security_policy": {
			Type:     types.BoolType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Description: "Disable PodSecurityPolicy in the API server and default manifests.",
		},
		"admission_control": {
			Optional:    true,
			Description: AdmissionPluginSchema.Description,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Attributes: tfsdk.ListNestedAttributes(AdmissionPluginSchema.Attributes),
		},
	},
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

// ProxyConfigSchema configures the Kubernetes control plane's kube-proxy.
var ProxyConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kube proxy configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
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

// TimeConfigSchema represents the options for configuring time on a machine.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#timeconfig for more information.
var TimeConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the options for configuring time on a machine.",
	Attributes: map[string]tfsdk.Attribute{
		"disabled": {
			Required:    true,
			Type:        types.StringType,
			Description: "Indicates if the time service is disabled for the machine. Defaults to false.",
		},
		"servers": {
			Optional: true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "Specifies time (NTP) servers to use for setting the system time. Defaults to pool.ntp.org",
		},
		"boot_timeout": {
			Optional: true,
			Type:     types.StringType,
			Description: `Specifies the timeout when the node time is considered to be in sync unlocking the boot sequence.
NTP sync will be still running in the background.
Defaults to “infinity” (waiting forever for time sync)`,
		},
	},
}

// LoggingConfigSchema configures Talos logging.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#loggingconfig for more information.
var LoggingConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures Talos logging.",
	Attributes: map[string]tfsdk.Attribute{
		"destinations": {
			Required:    true,
			Description: LoggingDestinationSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(LoggingDestinationSchema.Attributes),
		},
	},
}

// LoggingDestinationSchema configures Talos logging destination.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#loggingdestination for more information.
var LoggingDestinationSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures Talos logging destination.",
	Attributes: map[string]tfsdk.Attribute{
		"endpoint": {
			Required:    true,
			Description: "Where to send logs. Supported protocols are “tcp” and “udp”.",
			Type:        types.StringType,
		},
		"format": {
			Required:    true,
			Description: "Logs format.",
			Type:        types.StringType,
		},
	},
}

// KernelConfigSchema configures Talos Linux kernel.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#kernelconfig for more information.
var KernelConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures Talos Linux kernel.",
	Attributes: map[string]tfsdk.Attribute{
		"modules": {
			Required: true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Description: "Configures Linux kernel modules to load.",
		},
	},
}

// ControllerManagerConfigSchema represents the kube controller manager configuration options.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#controllermanagerconfig for more information.
var ControllerManagerConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kube controller manager configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Description: "The container image used in the controller manager manifest.",
		},
		"extra_args": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Extra arguments to supply to the controller manager.",
		},
		"extra_volumes": {
			Optional:    true,
			Description: VolumeMountSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(VolumeMountSchema.Attributes),
		},
		"env": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "The env field allows for the addition of environment variables for the control plane component.",
		},
	},
}

// SchedulerConfigSchema represents the kube scheduler configuration options.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#schedulerconfig for more information.
var SchedulerConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the kube scheduler configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Description: "The container image used in the scheduler manifest.",
		},
		"extra_args": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Extra arguments to supply to the scheduler.",
		},
		"extra_volumes": {
			Optional:    true,
			Description: VolumeMountSchema.Description,
			Attributes:  tfsdk.ListNestedAttributes(VolumeMountSchema.Attributes),
		},
		"env": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "The env field allows for the addition of environment variables for the control plane component.",
		},
	},
}

// ClusterDiscoveryConfigSchema configures cluster membership discovery.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#clusterdiscoveryconfig for more information.
var ClusterDiscoveryConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures cluster membership discovery.",
	Attributes: map[string]tfsdk.Attribute{
		"enabled": {
			Optional:    true,
			Description: "Enable cluster membership discovery",
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Type: types.BoolType,
		},
		"registries": {
			Optional:    true,
			Description: DiscoveryRegistriesConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(DiscoveryRegistriesConfigSchema.Attributes),
		},
	},
}

var DiscoveryRegistriesConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures cluster membership discovery.",
	Attributes: map[string]tfsdk.Attribute{
		"kubernetes_disabled": {
			Type:                types.BoolType,
			Required:            true,
			MarkdownDescription: "Disable Kubernetes discovery registry.",
		},
		"service_disabled": {
			Type:                types.BoolType,
			Required:            true,
			MarkdownDescription: "Disable external service discovery registry.",
		},
		// TODO make the kubernetes and service options mutually exclusive or pull them into their
		// own schema to be in line with the underlying datastructure.
		"service_endpoint": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "External service endpoint.",
		},
	},
}

// EtcdConfigSchema represents the etcd configuration options.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#etcdconfig for more information.
var EtcdConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the etcd configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			Description: "The container image used to create the etcd service.",
		},
		"ca_crt": {
			Type:                types.StringType,
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "PEM encoded etcd root certificate authority crt.",
		},
		"ca_key": {
			Type:                types.StringType,
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "PEM encoded etcd root certificate authority key.",
		},
		"extra_args": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Extra arguments to supply to etcd.",
		},
		"subnet": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "The subnet from which the advertise URL should be.",
		},
	},
}

// CoreDNSConfigSchema represents the CoreDNS config values.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#coredns for more information.
var CoreDNSConfigSchema tfsdk.Schema = tfsdk.Schema{
	MarkdownDescription: `Represents the CoreDNS config values.
Refer to [CoreDNS in the TalosOS Documentation](https://www.talos.dev/v1.0/reference/configuration/#coredns) for more information.`,
	Attributes: map[string]tfsdk.Attribute{
		"disabled": {
			Type:        types.BoolType,
			Required:    true,
			Description: "Disable coredns deployment on cluster bootstrap.",
		},
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			MarkdownDescription: "The `image` field is an override to the default coredns image.",
		},
	},
}

// AdminKubeconfigConfigSchema contains admin kubeconfig settings.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#adminkubeconfigconfig for more information.
var AdminKubeconfigConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Contains admin kubeconfig settings.",
	Attributes: map[string]tfsdk.Attribute{
		"cert_lifetime": {
			Type:     types.StringType,
			Required: true,
			MarkdownDescription: `Admin kubeconfig certificate lifetime (default is 1 year).
Field format accepts any Go time.Duration format (‘1h’ for one hour, ‘10m’ for ten minutes).`,
		},
	},
}

// ControlPlaneConfigSchema provides options for configuring the Kubernetes control plane.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#controlplaneconfig for more information.
var ControlPlaneConfigSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents the control plane configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"endpoint": {
			Type:        types.StringType,
			Optional:    true,
			Description: "Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.",
			Computed:    true,
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

// InstallSchema represents installation options for Talos nodes.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#installconfig for more information.
var InstallSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Represents installation options for Talos nodes.",
	Attributes: map[string]tfsdk.Attribute{
		"disk": {
			Type:     types.StringType,
			Optional: true,
		},
		"image": {
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
			// TODO validate
			// ValidateFunc: validateImage,
		},
		"bootloader": {
			Type:     types.BoolType,
			Optional: true,
		},
		"wipe": {
			Type:     types.BoolType,
			Optional: true,
		},
		"legacy_bios": {
			Type:     types.BoolType,
			Optional: true,
		},
		"extensions": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"kernel_args": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
	},
}

// FileSchema describes a machine file and it's contents to be written onto the node's filesystem.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#installconfig for more information.
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

// InlineManifestSchema describes inline bootstrap manifests for the user.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#clusterinlinemanifest for more information.
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

var CertBundleSchema = tfsdk.Schema{
	MarkdownDescription: "Represents the keys and certificates throughout Talos.",
	Attributes: map[string]tfsdk.Attribute{
		"admin_crt": {
			Optional:            true,
			Type:                types.StringType,
			MarkdownDescription: "PEM encoded cluster admin crt.",
		},
		"admin_key": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "PEM encoded cluster admin key.",
		},
		"etcd_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded etcd crt.",
		},
		"etcd_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded etcd key.",
		},
		"k8s_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded crt for k8s..",
		},
		"k8s_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for k8s.",
		},
		"k8s_aggregator_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded crt for the k8s aggregator.",
		},
		"k8s_aggregator_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for the k8s aggregator.",
		},
		"k8s_service_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for the k8s service.",
		},
		"os_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded crt for OS.",
		},
		"os_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for OS.",
		},
	},
}

var NetworkConfigOptionSchema = tfsdk.Schema{
	MarkdownDescription: "Represents globally applied network configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"with_kubespan": {
			Type:     types.BoolType,
			Optional: true,
		},
		"with_vip": {
			Optional:    true,
			Description: "Configures an interface for Virtual shared IP. Maps interfaces names to desired CIDRs.",
			Type: types.MapType{
				ElemType: types.StringType,
			},
		},
		"with_wireguard": {
			Optional:    true,
			Description: WireguardSchema.Description,
			Attributes:  tfsdk.MapNestedAttributes(WireguardSchema.Attributes),
		},
		"with_mtu": {
			Optional:    true,
			Description: "Configures an interface's MTU.",
			Type: types.MapType{
				ElemType: types.Int64Type,
			},
		},
		"with_cidr": {
			Optional:    true,
			Description: "Configures an interface for static addressing.",
			Type: types.MapType{
				ElemType: types.StringType,
			},
		},
		"with_dhcpv6": {
			Optional:    true,
			Description: "Enables DHCPv6 for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_dhcpv4": {
			Optional:    true,
			Description: "Enables DHCPv4 for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_dhcp": {
			Optional:    true,
			Description: "Enables DHCP for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_ignore": {
			Optional:    true,
			Description: "Enables DHCP for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_nameservers": {
			Optional:    true,
			Description: "Sets global nameservers list.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
		},
		"with_networkconfig": {
			Optional:    true,
			Description: NetworkConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(NetworkConfigSchema.Attributes),
		},
	},
}

var SecretBundleSchema = tfsdk.Schema{
	MarkdownDescription: "Represents secrets used throughout a Talos install.",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Unique cluster ID for Talos. Base64 encoded binary data.",
		},
		"cert_bundle": {
			Optional:            true,
			MarkdownDescription: CertBundleSchema.MarkdownDescription,
			Attributes:          tfsdk.SingleNestedAttributes(CertBundleSchema.Attributes),
		},
		"secret": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique cluster secret for Talos. Base64 encoded binary data.",
		},
		"bootstrap_token": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique token for Talos bootstrap.",
		},
		"aes_cbc_encryption": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique secret for Talos disk encryption. Base64 encoded binary data.",
		},
		"trustd_token": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique token for Talos trustd.",
		},
	},
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

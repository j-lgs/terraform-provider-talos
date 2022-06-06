package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/constants"
)

// Examples
var (
	RegistryExample = &Registry{
		Configs: map[string]RegistryConfig{"registry.local": *RegistryConfigExample},
		Mirrors: map[string][]types.String{
			"docker.io": Wrapsl(mirrorendpointsExample...),
		},
	}

	RegistryConfigExample = &RegistryConfig{
		Username:           Wraps(usernameExample),
		Password:           Wraps(passwordExample),
		Auth:               Wraps(authExample),
		IdentityToken:      Wraps(idtokenExample),
		ClientCRT:          Wraps(tlsCrtExample),
		ClientKey:          Wraps(tlsKeyExample),
		CA:                 Wraps(tlsCaExample),
		InsecureSkipVerify: Wrapb(tlsInsecureExample),
	}

	CniExample = &CNI{
		Name: Wraps(constants.CustomCNI),
		URLs: Wrapsl(cniURLsExample...),
	}

	KubeletExample = &KubeletConfig{
		Image:      Wraps((&v1alpha1.KubeletConfig{}).Image()),
		ClusterDNS: Wrapsl(ClusterDNSExample...),
		ExtraArgs: map[string]types.String{
			"feature-gates": Wraps("serverSideApply=true"),
		},
		ExtraMounts:        []ExtraMount{*ExtraMountExample},
		ExtraConfig:        Wraps(kubeletExtraConfigExampleString),
		RegisterWithFQDN:   Wrapb(kubeletRegisterWithFQDNExample),
		NodeIPValidSubnets: Wrapsl(kubeletSubnetExample...),
	}

	ExtraMountExample = &ExtraMount{
		Source:      Wraps(KubeletMountExample.Source),
		Destination: Wraps(KubeletMountExample.Destination),
		Type:        Wraps(KubeletMountExample.Type),
		Options:     Wrapsl(KubeletMountExample.Options...),
	}

	VolumeMountExample = &VolumeMount{
		HostPath:  Wraps(hostPathExample),
		MountPath: Wraps(mountPathExample),
		Readonly:  Wrapb(readOnlyExample),
	}

	InstallExample = &InstallConfig{
		Disk:       Wraps(installDiskExample),
		KernelArgs: Wrapsl(installKernelArgsExample...),
		Image:      Wraps(installImageExample),
		Extensions: Wrapsl(extensionImageExample),
		LegacyBios: Wrapb(installBiosExample),
		Bootloader: Wrapb(installBootloaderExample),
		Wipe:       Wrapb(installWipeExample),
	}

	NetworkConfigExample = &NetworkConfig{
		Hostname: Wraps(hostnameExample),
		Devices: []NetworkDevice{
			*WireguardDeviceExample,
			*StaticDeviceExample,
			*DummyDeviceExample1,
			*DummyDeviceExample2,
			*DummyDeviceExample3,
			*DummyDeviceExample4,
			*ActiveBackupBondDeviceExample,
			*LACPBondDeviceExample,
			*IgnoreDeviceExample,
			*VIPDeviceExample,
			*VLANDeviceExample,
			*DHCPDeviceExample,
		},
		Nameservers: Wrapsl(nameserversExample...),
		ExtraHosts: map[string][]types.String{
			extraHostExampleKey: Wrapsl(extraHostExampleValues...),
		},
		Kubespan: &NetworkKubeSpanExample,
	}

	WireguardDeviceExample = &NetworkDevice{
		Name:      Wraps("wg0"),
		Addresses: Wrapsl(wgDeviceExampleAddresses...),
		Wireguard: WireguardExample,
	}

	StaticDeviceExample = &NetworkDevice{
		Name:      Wraps("eth0"),
		Addresses: Wrapsl(staticAddressesExample...),
		Routes:    RoutesExample,
		MTU:       Wrapi(mtuExample),
	}

	DHCPDeviceExample = &NetworkDevice{
		Name: Wraps("eth8"),
		DHCP: Wrapb(true),
	}

	VIPDeviceExample = &NetworkDevice{
		Name: Wraps("eth7"),
		VIP:  VIPExample,
	}

	VLANDeviceExample = &NetworkDevice{
		Name: Wraps("eth6"),
		VLANs: []VLAN{
			*VLANExample,
		},
	}

	IgnoreDeviceExample = &NetworkDevice{
		Name:   Wraps("eth5"),
		Ignore: Wrapb(true),
	}

	DummyDeviceExample1 = &NetworkDevice{
		Name:  Wraps("eth1"),
		Dummy: Wrapb(true),
	}

	DummyDeviceExample2 = &NetworkDevice{
		Name:  Wraps("eth2"),
		Dummy: Wrapb(true),
	}

	DummyDeviceExample3 = &NetworkDevice{
		Name:  Wraps("eth3"),
		Dummy: Wrapb(true),
	}

	DummyDeviceExample4 = &NetworkDevice{
		Name:  Wraps("eth4"),
		Dummy: Wrapb(true),
	}

	ActiveBackupBondDeviceExample = &NetworkDevice{
		Name:     Wraps("bond0"),
		BondData: ActiveBackupBondDataExample,
	}

	LACPBondDeviceExample = &NetworkDevice{
		Name:     Wraps("bond1"),
		BondData: LACPBondDataExample,
	}

	ActiveBackupBondDataExample = &BondData{
		Interfaces:      Wrapsl(abBondInterfaceExample...),
		Mode:            Wraps(abBondModeExample),
		Primary:         Wraps(abPrimaryExample),
		PrimaryReselect: Wraps(abPrimaryReselectExample),
		ArpValidate:     Wraps(abArpValidateExample),
		//TlbDynamicLb:    Wrapi(abLacpTLBExample),
	}

	LACPBondDataExample = &BondData{
		Interfaces: Wrapsl(abLacpInterfacesExample...),
		// Unsupported
		//ARPIPTarget:    Wrapsl("0.0.0.0"),
		Mode:           Wraps(abLacpModeExample),
		XmitHashPolicy: Wraps(abLacpXmitExample),
		LacpRate:       Wraps(abLacpRateExample),
		// Unsupported
		//AdActorSystem:  Wraps("00:00:00:00:00:
		ArpAllTargets: Wraps(abLacpArpAllExample),
		FailoverMac:   Wraps(abLacpFailoverMacExample),
		AdSelect:      Wraps(abLacpADSelectExample),
		MiiMon:        Wrapi(abLacpMiimonExample),
		UpDelay:       Wrapi(abLacpUpDelayExample),
		DownDelay:     Wrapi(abLacpDownDelayExample),
		//ArpInterval:     Wrapi(abLacpArpIntervalExample),
		ResendIgmp: Wrapi(abLacpResendIgmpExample),
		MinLinks:   Wrapi(abLacpMinLinksExample),
		LpInterval: Wrapi(abLacpLPIntervalExample),
		//PacketsPerSlave: Wrapi(abLacpBondPacketsPerExample),
		NumPeerNotif:    Wrapi(abLacpNumPeerExample),
		AllSlavesActive: Wrapi(abLacpAllSlavesExample),
		// TODO Does not set properly
		//UseCarrier:      Wrapb(*abLacpUseCarrierExample),
		AdActorSysPrio:  Wrapi(abLacpAdActorExample),
		AdUserPortKey:   Wrapi(abLacpUserPortExample),
		PeerNotifyDelay: Wrapi(abLacpPeerNotifDelayExample),
	}

	DHCPOptionsExample = &DHCPOptions{
		RouteMetric: Wrapi(dhcpMetricExample),
		IPV4:        Wrapb(dhcpIpv4Example),
		IPV6:        Wrapb(dhcpIpv6Example),
	}

	VLANExample = &VLAN{
		Addresses: Wrapsl(vlanAddressesExample...),
		Routes:    RoutesExample,
		DHCP:      Wrapb(vlanDHCPExample),
		VLANId:    Wrapi(vlanIDExample),
		MTU:       Wrapi(vlanMTUExample),
		VIP:       VIPExample,
	}

	RoutesExample = []Route{
		{
			Network: Wraps(routeNetworkExample),
			Gateway: Wraps(routeGatewayExample),
		}, {
			Network: Wraps(altRouteNetworkExample),
			Gateway: Wraps(altGatewayExample),
			Source:  Wraps(routeSourceExample),
			Metric:  Wrapi(routeMetricExample),
		},
	}

	WireguardExample = &Wireguard{
		Peers:        []WireguardPeer{WireguardPeerExample},
		FirewallMark: Wrapi(wgFirewallMarkExample),
		ListenPort:   Wrapi(wgListenPortExample),
		PublicKey:    Wraps(wgPublicKeyExample),
		PrivateKey:   Wraps(wgPrivateKeyExample),
	}

	WireguardPeerExample = WireguardPeer{
		AllowedIPs:                  Wrapsl(wgAllowedIPsExample...),
		Endpoint:                    Wraps(wgEndpointExample),
		PersistentKeepaliveInterval: Wrapi(wgPersistentKeepaliveExample),
		PublicKey:                   Wraps(wgPublicKeyExample),
	}

	VIPExample = &VIP{
		IP:                   Wraps(vipCIDRExample),
		EquinixMetalAPIToken: Wraps(vipTokenExample),
		HetznerCloudAPIToken: Wraps(vipTokenExample),
	}

	MachineControlPlaneExample = &MachineControlPlane{
		ControllerManagerDisabled: Wrapb(ControllerManagerDisabledExample),
		SchedulerDisabled:         Wrapb(SchedulerDisabledExample),
	}

	MachineDiskExample = &MachineDiskData{
		DeviceName: Wraps(deviceNameExample),
		Partitions: []PartitionData{*PartitionDataExample},
	}

	PartitionDataExample = &PartitionData{
		Size:       Wraps(diskSizeExampleStr),
		MountPoint: Wraps(diskMountExample),
	}

	EncryptionDataExample = &EncryptionData{
		State:     EncryptionConfigSystemExample,
		Ephemeral: EncryptionConfigEphemeralExample,
	}

	EncryptionConfigSystemExample = &EncryptionConfigData{
		Provider:    Wraps(providerExample),
		Keys:        []KeyConfig{KeyConfigStaticExample},
		Cipher:      Wraps(cipherExample),
		KeySize:     Wrapi(keysizeExample),
		BlockSize:   Wrapi(blocksizeExample),
		PerfOptions: Wrapsl(perfoptsExample...),
	}

	EncryptionConfigEphemeralExample = &EncryptionConfigData{
		Provider:    Wraps(providerExample),
		Keys:        []KeyConfig{KeyConfigNodeIDExample},
		Cipher:      Wraps(cipherExample),
		KeySize:     Wrapi(keysizeExample),
		BlockSize:   Wrapi(blocksizeExample),
		PerfOptions: Wrapsl(perfoptsExample...),
	}

	KeyConfigStaticExample = KeyConfig{
		NodeID:    types.Bool{Null: true},
		KeyStatic: Wraps(keydataExample),
		Slot:      Wrapi(0),
	}

	KeyConfigNodeIDExample = KeyConfig{
		NodeID:    Wrapb(true),
		KeyStatic: types.String{Null: true},
		Slot:      Wrapi(0),
	}

	TimeConfigExample = &TimeConfig{
		Disabled:    Wrapb(timedisabledExample),
		Servers:     Wrapsl(timeserversExample...),
		BootTimeout: Wraps(timeoutExampleString),
	}

	LoggingConfigExample = &LoggingConfig{
		Destinations: []LoggingDestination{
			*LoggingDestinationExample,
		},
	}

	LoggingDestinationExample = &LoggingDestination{
		Endpoint: Wraps(loggingEndpointExample.String()),
		Format:   Wraps(loggingFormatExample),
	}

	KernelConfigExample = &KernelConfig{
		Modules: Wrapsl(kernelModuleNameExample),
	}

	ProxyConfigExample = &ProxyConfig{
		Image:    Wraps((&v1alpha1.ProxyConfig{}).Image()),
		Mode:     Wraps(proxyModeExample),
		Disabled: Wrapb(proxyDisabledExample),
		ExtraArgs: map[string]types.String{
			"proxy-mode": Wraps("iptables"),
		},
	}

	ControlPlaneConfigExample = &ControlPlaneConfig{
		Endpoint:           Wraps(EndpointExample.String()),
		LocalAPIServerPort: Wrapi(localApiserverExample),
	}

	NetworkKubeSpanExample = NetworkKubeSpan{
		Enabled:             Wrapb(true),
		AllowPeerDownBypass: Wrapb(true),
	}

	ControllerManagerExample = &ControllerManagerConfig{
		Image:        s((&v1alpha1.ControllerManagerConfig{}).Image()),
		ExtraArgs:    controllerManagerExtraArgsTFExample,
		ExtraVolumes: []VolumeMount{*VolumeMountExample},
		Env:          controllerManagerEnvTFExample,
	}

	ClusterDiscoveryConfigExample = &ClusterDiscoveryConfig{
		Enabled:    Wrapb(discoveryExample),
		Registries: DiscoveryRegistriesConfigExample,
	}

	DiscoveryRegistriesConfigExample = &DiscoveryRegistriesConfig{
		KubernetesDisabled: Wrapb(discoveryRegistryKubernetesEnabledExample),
		ServiceDisabled:    Wrapb(discoveryRegistryServiceEnabledExample),
		ServiceEndpoint:    Wraps(constants.DefaultDiscoveryServiceEndpoint),
	}

	EtcdConfigExample = &EtcdConfig{
		Image:     Wraps((&v1alpha1.EtcdConfig{}).Image()),
		CaCrt:     Wraps(string(etcdCertsExample.Crt)),
		CaKey:     Wraps(string(etcdCertsExample.Key)),
		ExtraArgs: etcdArgsTFExample,
		Subnet:    Wraps(etcdSubnetExample),
	}

	SchedulerExample = &SchedulerConfig{
		Image:        Wraps((&v1alpha1.SchedulerConfig{}).Image()),
		ExtraArgs:    schedulerArgsTFExample,
		ExtraVolumes: []VolumeMount{*VolumeMountExample},
		Env:          schedulerEnvTFExample,
	}

	CoreDNSExample = &CoreDNS{
		Image:    Wraps((&v1alpha1.CoreDNS{}).Image()),
		Disabled: Wrapb(coreDNSDisabledExample),
	}

	AdminKubeconfigConfigExample = &AdminKubeconfigConfig{
		CertLifetime: Wraps("8760h"),
	}

	APIServerExample = &APIServerConfig{
		Image: s((&v1alpha1.APIServerConfig{}).Image()),
		ExtraArgs: map[string]types.String{
			"feature-gates":                    Wraps("ServerSideApply=true"),
			"http2-max-streams-per-connection": Wraps("32"),
		},
		ExtraVolumes: []VolumeMount{*VolumeMountExample},
		Env: map[string]types.String{
			"key": Wraps("value"),
		},
		CertSANS:         Wrapsl(apiServerSANsExample...),
		DisablePSP:       Wrapb(apiServerDisablePSPExample),
		AdmissionPlugins: []AdmissionPluginConfig{AdmissionPluginExample},
	}

	AdmissionPluginExample = AdmissionPluginConfig{
		Name:          Wraps(pluginNameExample),
		Configuration: Wraps(pluginConfigExample),
	}

	FileExample = File{
		Content:     Wraps(machineFileContentExample),
		Permissions: Wrapi(machineFilePermissionsExample),
		Path:        Wraps(machineFilePathExample),
		Op:          Wraps(machineFileOpExample),
	}

	InlineManifestExample = InlineManifest{
		Name:    Wraps(inlineManifestNameExample),
		Content: Wraps(inlineManifestContentExample),
	}
)

type InstallConfig struct {
	Disk       types.String   `tfsdk:"disk"`
	KernelArgs []types.String `tfsdk:"kernel_args"`
	Image      types.String   `tfsdk:"image"`
	Bootloader types.Bool     `tfsdk:"bootloader"`
	LegacyBios types.Bool     `tfsdk:"legacy_bios"`
	Extensions []types.String `tfsdk:"extensions"`
	Wipe       types.Bool     `tfsdk:"wipe"`
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

// Registry represents the image pull options.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#registriesconfig for more information.
type Registry struct {
	Mirrors map[string][]types.String `tfsdk:"mirrors"`
	Configs map[string]RegistryConfig `tfsdk:"configs"`
}

type CNI struct {
	Name types.String   `tfsdk:"name"`
	URLs []types.String `tfsdk:"urls"`
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

// ExtraMount wraps the OCI mount specification.
// Refer to https://github.com/opencontainers/runtime-spec/blob/main/config.md#mounts for more information.
type ExtraMount struct {
	Destination types.String   `tfsdk:"destination"`
	Type        types.String   `tfsdk:"type"`
	Source      types.String   `tfsdk:"source"`
	Options     []types.String `tfsdk:"options"`
}

// VolumeMount Describes extra volume mounts for controlplane static pods.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#volumemountconfig for more information.
type VolumeMount struct {
	HostPath  types.String `tfsdk:"host_path"`
	MountPath types.String `tfsdk:"mount_path"`
	Readonly  types.Bool   `tfsdk:"readonly"`
}

type NetworkConfig struct {
	Hostname    types.String              `tfsdk:"hostname"`
	Devices     []NetworkDevice           `tfsdk:"devices"`
	Nameservers []types.String            `tfsdk:"nameservers"`
	ExtraHosts  map[string][]types.String `tfsdk:"extra_hosts"`
	Kubespan    *NetworkKubeSpan          `tfsdk:"kubespan"`
}

// NetworkDevice describes a Talos Device configuration.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#device for more information.
// TODO: Add network device selector field for interfaces and support it throughout the provider.
type NetworkDevice struct {
	Name        types.String   `tfsdk:"name"`
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

// DHCPOptions specificies DHCP specific options.
type DHCPOptions struct {
	RouteMetric types.Int64 `tfsdk:"route_metric"`
	IPV4        types.Bool  `tfsdk:"ipv4"`
	IPV6        types.Bool  `tfsdk:"ipv6"`
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

// Route represents a network route.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#route for more information.
type Route struct {
	Network types.String `tfsdk:"network"`
	Gateway types.String `tfsdk:"gateway"`
	Source  types.String `tfsdk:"source"`
	Metric  types.Int64  `tfsdk:"metric"`
}

// Wireguard describes a network interface's Wireguard configuration and keys.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#devicewireguardconfig for more information.
type Wireguard struct {
	Peers        []WireguardPeer `tfsdk:"peer"`
	FirewallMark types.Int64     `tfsdk:"firewall_mark"`
	ListenPort   types.Int64     `tfsdk:"listen_port"`
	PublicKey    types.String    `tfsdk:"public_key"`
	PrivateKey   types.String    `tfsdk:"private_key"`
}

// WireguardPeer describes an interface's Wireguard peers.
type WireguardPeer struct {
	AllowedIPs                  []types.String `tfsdk:"allowed_ips"`
	Endpoint                    types.String   `tfsdk:"endpoint"`
	PersistentKeepaliveInterval types.Int64    `tfsdk:"persistent_keepalive_interval"`
	PublicKey                   types.String   `tfsdk:"public_key"`
}

// VIP represent virtual shared IP configurations for network interfaces.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#devicevipconfig for more information.
type VIP struct {
	IP                   types.String `tfsdk:"ip"`
	EquinixMetalAPIToken types.String `tfsdk:"equinix_metal_api_token"`
	HetznerCloudAPIToken types.String `tfsdk:"hetzner_cloud_api_token"`
}

// MachineControlPlane configures options pertaining to the Kubernetes control plane that's installed onto the machine.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#machinecontrolplaneconfig for more information.
type MachineControlPlane struct {
	ControllerManagerDisabled types.Bool `tfsdk:"controller_manager_disabled"`
	SchedulerDisabled         types.Bool `tfsdk:"scheduler_disabled"`
}

type MachineDiskDataList []MachineDiskData

// MachineDiskData represents the options available for partitioning, formatting, and mounting extra disks.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#machinedisk for more information.
type MachineDiskData struct {
	DeviceName types.String    `tfsdk:"device_name"`
	Partitions []PartitionData `tfsdk:"partitions"`
}

// PartitionData represents the options for a disk partition.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#diskpartition for more information.
type PartitionData struct {
	Size       types.String `tfsdk:"size"`
	MountPoint types.String `tfsdk:"mount_point"`
}

type SysctlData map[string]types.String

// EncryptionData specifies system disk partitions encryption settings.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#systemdiskencryptionconfig for more information.
type EncryptionData struct {
	State     *EncryptionConfigData `tfsdk:"state"`
	Ephemeral *EncryptionConfigData `tfsdk:"ephemeral"`
}

// EncryptionConfigData represents partition encryption settings.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#encryptionconfig for more information.
type EncryptionConfigData struct {
	Provider    types.String   `tfsdk:"crypt_provider"`
	Keys        []KeyConfig    `tfsdk:"keys"`
	Cipher      types.String   `tfsdk:"cipher"`
	KeySize     types.Int64    `tfsdk:"keysize"`
	BlockSize   types.Int64    `tfsdk:"blocksize"`
	PerfOptions []types.String `tfsdk:"perf_options"`
}

// KeyConfig represents configuration for disk encryption key.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#encryptionkey for more information.
type KeyConfig struct {
	KeyStatic types.String `tfsdk:"key_static"`
	NodeID    types.Bool   `tfsdk:"node_id"`
	Slot      types.Int64  `tfsdk:"slot"`
}

// ProxyConfig configures the Kubernetes control plane's kube-proxy.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#proxyconfig for more information.
type ProxyConfig struct {
	Image     types.String            `tfsdk:"image"`
	Mode      types.String            `tfsdk:"mode"`
	Disabled  types.Bool              `tfsdk:"is_disabled"`
	ExtraArgs map[string]types.String `tfsdk:"extra_args"`
}

// ControllerManagerConfig represents the kube controller manager configuration options.
type ControllerManagerConfig struct {
	Image        types.String            `tfsdk:"image"`
	ExtraArgs    map[string]types.String `tfsdk:"extra_args"`
	ExtraVolumes []VolumeMount           `tfsdk:"extra_volumes"`
	Env          map[string]types.String `tfsdk:"env"`
}

// SchedulerConfig represents the kube scheduler configuration options.
type SchedulerConfig struct {
	Image        types.String            `tfsdk:"image"`
	ExtraArgs    map[string]types.String `tfsdk:"extra_args"`
	ExtraVolumes []VolumeMount           `tfsdk:"extra_volumes"`
	Env          map[string]types.String `tfsdk:"env"`
}

// ControlPlaneConfig provides options for configuring the Kubernetes control plane.
// Refer to https://www.talos.dev/v1.0/reference/configuration/#controlplaneconfig for more information.
type ControlPlaneConfig struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	LocalAPIServerPort types.Int64  `tfsdk:"local_api_server_port"`
}

// EtcdConfig represents the etcd configuration options.
type EtcdConfig struct {
	Image     types.String            `tfsdk:"image"`
	ExtraArgs map[string]types.String `tfsdk:"extra_args"`
	CaCrt     types.String            `tfsdk:"ca_crt"`
	CaKey     types.String            `tfsdk:"ca_key"`
	Subnet    types.String            `tfsdk:"subent"`
}

// ClusterDiscoveryConfig struct configures cluster membership discovery.
type ClusterDiscoveryConfig struct {
	Enabled    types.Bool                 `tfsdk:"enabled"`
	Registries *DiscoveryRegistriesConfig `tfsdk:"registries"`
}

// DiscoveryRegistriesConfig struct configures cluster membership discovery.
type DiscoveryRegistriesConfig struct {
	KubernetesDisabled types.Bool   `tfsdk:"kubernetes_disabled"`
	ServiceDisabled    types.Bool   `tfsdk:"service_disabled"`
	ServiceEndpoint    types.String `tfsdk:"service_endpoint"`
}

// CoreDNS represents the CoreDNS config values.
type CoreDNS struct {
	Image    types.String `tfsdk:"image"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

// AdminKubeconfigConfig contains admin kubeconfig settings.
type AdminKubeconfigConfig struct {
	CertLifetime types.String `tfsdk:"cert_lifetime"`
}

// NetworkKubeSpan describes KubeSpan configuration
// Refer to https://www.talos.dev/v1.0/reference/configuration/#networkkubespan for more information.
type NetworkKubeSpan struct {
	Enabled             types.Bool `tfsdk:"enabled"`
	AllowPeerDownBypass types.Bool `tfsdk:"allow_peer_down_bypass"`
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

// AdmissionPluginConfig configures pod admssion rules on the kubelet64Type, denying execution to pods that don't fit them.
type AdmissionPluginConfig struct {
	Name          types.String `tfsdk:"name"`
	Configuration types.String `tfsdk:"configuration"`
}

// File describes a machine file and it's contents to be written onto the node's filesystem.
type File struct {
	Content     types.String `tfsdk:"content"`
	Permissions types.Int64  `tfsdk:"permissions"`
	Path        types.String `tfsdk:"path"`
	Op          types.String `tfsdk:"op"`
}

// TimeConfig represents the options for configuring time on a machine.
type TimeConfig struct {
	Disabled    types.Bool     `tfsdk:"disabled"`
	Servers     []types.String `tfsdk:"servers"`
	BootTimeout types.String   `tfsdk:"boot_timeout"`
}

// LoggingConfig configures Talos logging and its destinations.
type LoggingConfig struct {
	Destinations []LoggingDestination `tfsdk:"destinations"`
}

type LoggingDestination struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Format   types.String `tfsdk:"format"`
}

// KernelConfig configures the Talos Linux kernel.
type KernelConfig struct {
	Modules []types.String `tfsdk:"modules"`
}

// InlineManifest describes inline bootstrap manifests for the user.
type InlineManifest struct {
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
}

type CertBundle struct {
	AdminCRT         types.String `tfsdk:"admin_crt"`
	AdminKey         types.String `tfsdk:"admin_key"`
	EtcdCRT          types.String `tfsdk:"etcd_crt"`
	EtcdKey          types.String `tfsdk:"etcd_key"`
	K8sCRT           types.String `tfsdk:"k8s_crt"`
	K8sKey           types.String `tfsdk:"k8s_key"`
	K8sAggregatorCRT types.String `tfsdk:"k8s_aggregator_crt"`
	K8sAggregatorKey types.String `tfsdk:"k8s_aggregator_key"`
	K8sServiceKey    types.String `tfsdk:"k8s_service_key"`
	OSCRT            types.String `tfsdk:"os_crt"`
	OSKey            types.String `tfsdk:"os_key"`
}

type NetworkConfigOptions struct {
	Kubespan      types.Bool              `tfsdk:"with_kubespan"`
	VIP           map[string]types.String `tfsdk:"with_vip"`
	Wireguard     map[string]Wireguard    `tfsdk:"with_wireguard"`
	MTU           map[string]types.Int64  `tfsdk:"with_mtu"`
	CIDR          map[string]types.String `tfsdk:"with_cidr"`
	DHCPv6        map[string]types.Bool   `tfsdk:"with_dhcpv6"`
	DHCPv4        map[string]types.Bool   `tfsdk:"with_dhcpv4"`
	DHCP          map[string]types.Bool   `tfsdk:"with_dhcp"`
	Ignore        map[string]types.Bool   `tfsdk:"with_ignore"`
	Nameservers   []types.String          `tfsdk:"with_nameservers"`
	NetworkConfig *NetworkConfig          `tfsdk:"with_networkconfig"`
}

type SecretBundle struct {
	ID             types.String `tfsdk:"id"`
	CertBundle     *CertBundle  `tfsdk:"cert_bundle"`
	Secret         types.String `tfsdk:"secret"`
	BootstrapToken types.String `tfsdk:"bootstrap_token"`
	AESEncryption  types.String `tfsdk:"aes_cbc_encryption"`
	TrustdToken    types.String `tfsdk:"trustd_token"`
}

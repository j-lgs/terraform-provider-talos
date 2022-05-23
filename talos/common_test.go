package talos

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/opencontainers/runtime-spec/specs-go"
	sdiff "github.com/r3labs/diff/v3"
	"github.com/talos-systems/go-blockdevice/blockdevice/util/disk"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/constants"

	configloader "github.com/talos-systems/talos/pkg/machinery/config/configloader"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	genv1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
)

func s(str string) types.String {
	return types.String{Value: str}
}

func sl(strs ...string) (out []types.String) {
	for _, st := range strs {
		out = append(out, s(st))
	}
	return
}

func wrapi(i int) types.Int64 {
	return types.Int64{Value: int64(i)}
}

func wraps(s string) types.String {
	return types.String{Value: s}
}

func wrapb(b bool) types.Bool {
	return types.Bool{Value: b}
}

var (
	testFalse bool = false
	testTrue  bool = true

	date  = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	clock = genv1alpha1.NewClock()

	testBundle = genv1alpha1.SecretsBundle{
		Clock: clock,
		Cluster: &genv1alpha1.Cluster{
			ID:     "tSuqMd_jk2CU_wGDuPpE7A3HlY9_mcoXWWJ0kRbK8aE=",
			Secret: "6SxVdcxHbbUdSsPpgnnRSHClbxkwmVpxNnbIKVGVirk=",
		},
		Secrets:    testSecrets,
		TrustdInfo: testTrustdInfo,
		Certs:      testDefaultCerts,
	}

	expectedNode *v1alpha1.Config = &v1alpha1.Config{
		ConfigVersion: "v1alpha1",
		ConfigDebug:   testDebug,
		ConfigPersist: testPersist,
		MachineConfig: &v1alpha1.MachineConfig{
			MachineType:     "controlplane",
			MachineToken:    "s1pygp.a474wnneqo4v3lbs",
			MachineCA:       testBundle.Certs.OS,
			MachineCertSANs: testCertSANs,
			MachineControlPlane: &v1alpha1.MachineControlPlaneConfig{
				MachineControllerManager: &v1alpha1.MachineControllerManagerConfig{
					MachineControllerManagerDisabled: false,
				},
				MachineScheduler: &v1alpha1.MachineSchedulerConfig{
					MachineSchedulerDisabled: false,
				},
			},
			MachineKubelet: &v1alpha1.KubeletConfig{
				KubeletImage: "ghcr.io/siderolabs/kubelet:v1.23.6",
				KubeletClusterDNS: []string{
					"0.0.0.0",
				},
				KubeletExtraArgs: map[string]string{
					"arg": "value",
				},
				KubeletExtraMounts: []v1alpha1.ExtraMount{
					{
						Mount: specs.Mount{
							Destination: "/",
							Type:        "bind",
							Source:      "/",
							Options: []string{
								"noatime",
							},
						},
					},
				},
				KubeletExtraConfig: v1alpha1.Unstructured{
					Object: map[string]interface{}{
						"key": "value",
					},
				},
				KubeletRegisterWithFQDN: false,
				KubeletNodeIP: v1alpha1.KubeletNodeIPConfig{
					KubeletNodeIPValidSubnets: []string{
						"0.0.0.0/0",
					},
				},
			},
			MachinePods: []v1alpha1.Unstructured{
				{
					Object: map[string]interface{}{
						"key": "value",
					},
				},
			},
			MachineNetwork: &v1alpha1.NetworkConfig{
				NetworkHostname: "test-node",
				NetworkInterfaces: []*v1alpha1.Device{
					{
						DeviceInterface: "wg0",
						DeviceAddresses: []string{
							"0.0.0.0",
						},
						DeviceCIDR:        "",
						DeviceRoutes:      nil,
						DeviceBond:        nil,
						DeviceVlans:       nil,
						DeviceMTU:         0,
						DeviceDHCP:        false,
						DeviceIgnore:      false,
						DeviceDummy:       false,
						DeviceDHCPOptions: nil,
						DeviceWireguardConfig: &v1alpha1.DeviceWireguardConfig{
							WireguardPrivateKey:   "QDXghbrZoJ+NvLo9MOsaP4JARcPqa0Gy5lXV9EgtNWk=",
							WireguardListenPort:   0,
							WireguardFirewallMark: 0,
							WireguardPeers: []*v1alpha1.DeviceWireguardPeer{
								{
									WireguardPublicKey:                   "JbHCJXTOS6wRDjZM1an5YHxGz4QsU7VZKim5EBtpMxk=",
									WireguardEndpoint:                    "0.0.0.0:0",
									WireguardPersistentKeepaliveInterval: 25000000000,
									WireguardAllowedIPs: []string{
										"0.0.0.0/0",
									},
								},
							},
						},
						DeviceVIPConfig: nil,
					}, {
						DeviceInterface: "eth0",
						DeviceAddresses: []string{
							"0.0.0.0",
						},
						DeviceCIDR: "",
						DeviceRoutes: []*v1alpha1.Route{
							{
								RouteNetwork: "0.0.0.0/24",
								RouteGateway: "0.0.0.0",
								RouteSource:  "0.0.0.0",
								RouteMetric:  0,
							},
						},
						DeviceBond:            nil,
						DeviceVlans:           nil,
						DeviceMTU:             0,
						DeviceDHCP:            false,
						DeviceIgnore:          false,
						DeviceDummy:           false,
						DeviceDHCPOptions:     nil,
						DeviceWireguardConfig: nil,
						DeviceVIPConfig:       nil,
					}, {
						DeviceInterface:       "eth1",
						DeviceAddresses:       nil,
						DeviceCIDR:            "",
						DeviceRoutes:          nil,
						DeviceBond:            nil,
						DeviceVlans:           nil,
						DeviceMTU:             0,
						DeviceDHCP:            false,
						DeviceIgnore:          false,
						DeviceDummy:           false,
						DeviceDHCPOptions:     nil,
						DeviceWireguardConfig: nil,
						DeviceVIPConfig:       nil,
					}, {
						DeviceInterface:       "eth2",
						DeviceAddresses:       nil,
						DeviceCIDR:            "",
						DeviceRoutes:          nil,
						DeviceBond:            nil,
						DeviceVlans:           nil,
						DeviceMTU:             0,
						DeviceDHCP:            false,
						DeviceIgnore:          false,
						DeviceDummy:           false,
						DeviceDHCPOptions:     nil,
						DeviceWireguardConfig: nil,
						DeviceVIPConfig:       nil,
					}, {
						DeviceInterface:       "eth3",
						DeviceAddresses:       nil,
						DeviceCIDR:            "",
						DeviceRoutes:          nil,
						DeviceBond:            nil,
						DeviceVlans:           nil,
						DeviceMTU:             0,
						DeviceDHCP:            false,
						DeviceIgnore:          false,
						DeviceDummy:           false,
						DeviceDHCPOptions:     nil,
						DeviceWireguardConfig: nil,
						DeviceVIPConfig:       nil,
					}, {
						DeviceInterface:       "eth4",
						DeviceAddresses:       nil,
						DeviceCIDR:            "",
						DeviceRoutes:          nil,
						DeviceBond:            nil,
						DeviceVlans:           nil,
						DeviceMTU:             0,
						DeviceDHCP:            false,
						DeviceIgnore:          false,
						DeviceDummy:           false,
						DeviceDHCPOptions:     nil,
						DeviceWireguardConfig: nil,
						DeviceVIPConfig:       nil,
					}, {
						DeviceInterface: "bond0",
						DeviceAddresses: nil,
						DeviceCIDR:      "",
						DeviceRoutes:    nil,
						DeviceBond: &v1alpha1.Bond{
							BondInterfaces: []string{
								"eth1",
								"eth2",
							},
							BondARPIPTarget:     nil,
							BondMode:            "active-backup",
							BondHashPolicy:      "",
							BondLACPRate:        "",
							BondADActorSystem:   "",
							BondARPValidate:     "all",
							BondARPAllTargets:   "",
							BondPrimary:         "eth1",
							BondPrimaryReselect: "better",
							BondFailOverMac:     "",
							BondADSelect:        "",
							BondMIIMon:          0,
							BondUpDelay:         0,
							BondDownDelay:       0,
							BondARPInterval:     0,
							BondResendIGMP:      0,
							BondMinLinks:        0,
							BondLPInterval:      0,
							BondPacketsPerSlave: 0,
							BondNumPeerNotif:    0,
							BondTLBDynamicLB:    0,
							BondAllSlavesActive: 0,
							BondUseCarrier:      &testFalse,
							BondADActorSysPrio:  0,
							BondADUserPortKey:   0,
							BondPeerNotifyDelay: 0,
						},
						DeviceVlans:           nil,
						DeviceMTU:             0,
						DeviceDHCP:            false,
						DeviceIgnore:          false,
						DeviceDummy:           false,
						DeviceDHCPOptions:     nil,
						DeviceWireguardConfig: nil,
						DeviceVIPConfig:       nil,
					}, {
						DeviceInterface: "bond1",
						DeviceAddresses: nil,
						DeviceCIDR:      "",
						DeviceRoutes:    nil,
						DeviceBond: &v1alpha1.Bond{
							BondInterfaces: []string{
								"eth3",
								"eth4",
							},
							BondARPIPTarget:     nil,
							BondMode:            "802.3ad",
							BondHashPolicy:      "layer2+3",
							BondLACPRate:        "fast",
							BondADActorSystem:   "",
							BondARPValidate:     "",
							BondARPAllTargets:   "any",
							BondPrimary:         "",
							BondPrimaryReselect: "",
							BondFailOverMac:     "active",
							BondADSelect:        "stable",
							BondMIIMon:          100,
							BondUpDelay:         200,
							BondDownDelay:       200,
							BondARPInterval:     0,
							BondResendIGMP:      0,
							BondMinLinks:        0,
							BondLPInterval:      0,
							BondPacketsPerSlave: 0,
							BondNumPeerNotif:    0,
							BondTLBDynamicLB:    0,
							BondAllSlavesActive: 0,
							BondUseCarrier:      &testFalse,
							BondADActorSysPrio:  0,
							BondADUserPortKey:   0,
							BondPeerNotifyDelay: 0,
						},
						DeviceVlans: []*v1alpha1.Vlan{
							{
								VlanAddresses: []string{
									"0.0.0.0",
								},
								VlanCIDR: "",
								VlanRoutes: []*v1alpha1.Route{
									{
										RouteNetwork: "0.0.0.0/24",
										RouteGateway: "0.0.0.0",
										RouteSource:  "0.0.0.0",
										RouteMetric:  0,
									},
								},
								VlanDHCP: false,
								VlanID:   1,
								VlanMTU:  1500,
								VlanVIP: &v1alpha1.DeviceVIPConfig{
									SharedIP: "0.0.0.0",
									EquinixMetalConfig: &v1alpha1.VIPEquinixMetalConfig{
										EquinixMetalAPIToken: "token",
									},
									HCloudConfig: &v1alpha1.VIPHCloudConfig{
										HCloudAPIToken: "token",
									},
								},
							},
						},
						DeviceMTU:    1500,
						DeviceDHCP:   false,
						DeviceIgnore: false,
						DeviceDummy:  false,
						DeviceDHCPOptions: &v1alpha1.DHCPOptions{
							DHCPRouteMetric: 200,
							DHCPIPv4:        &testTrue,
							DHCPIPv6:        &testTrue,
						},
						DeviceWireguardConfig: nil,
						DeviceVIPConfig: &v1alpha1.DeviceVIPConfig{
							SharedIP: "0.0.0.0",
							EquinixMetalConfig: &v1alpha1.VIPEquinixMetalConfig{
								EquinixMetalAPIToken: "token",
							},
							HCloudConfig: &v1alpha1.VIPHCloudConfig{
								HCloudAPIToken: "token",
							},
						},
					},
				},
				NameServers: []string{
					"0.0.0.0",
				},
				ExtraHostEntries: []*v1alpha1.ExtraHost{
					{
						HostIP: "domain",
						HostAliases: []string{
							"0.0.0.0",
						},
					},
				},
				NetworkKubeSpan: v1alpha1.NetworkKubeSpan{
					KubeSpanEnabled:             true,
					KubeSpanAllowDownPeerBypass: true,
				},
			},
			MachineDisks: testMachineDisks,
			MachineInstall: &v1alpha1.InstallConfig{
				InstallDisk: testInstallDisk,
				InstallDiskSelector: &v1alpha1.InstallDiskSelector{
					Size: &v1alpha1.InstallDiskSizeMatcher{
						Matcher: disk.WithName("WDC*"),
					},
					Name:     "selector",
					Model:    "WDC*",
					Serial:   "serial",
					Modalias: "alias",
					UUID:     "UUID",
					WWID:     "nvme",
					Type:     v1alpha1.InstallDiskType(disk.TypeNVMe),
					BusPath:  "/pci0000:00/*",
				},
				InstallExtraKernelArgs: testInstallExtraKernelArgs,
				InstallImage:           testInstallImage,
				InstallExtensions: []v1alpha1.InstallExtensionConfig{
					{
						ExtensionImage: "extensions:latest",
					},
				},
				InstallBootloader:        false,
				InstallWipe:              true,
				InstallLegacyBIOSSupport: true,
			},
			MachineFiles: []*v1alpha1.MachineFile{
				{
					FileContent:     "file",
					FilePermissions: 420,
					FilePath:        "/path",
					FileOp:          "create",
				},
			},
			MachineEnv: map[string]string{
				"var": "value",
			},
			MachineTime: &v1alpha1.TimeConfig{
				TimeDisabled: true,
				TimeServers: []string{
					"time.test",
				},
				TimeBootTimeout: time.Minute * 1,
			},
			MachineSysctls: testSysctls,
			MachineSysfs: map[string]string{
				"key": "value",
			},
			MachineRegistries: v1alpha1.RegistriesConfig{
				RegistryMirrors: testRegistryMirror,
				RegistryConfig:  testRegistryConfig,
			},
			MachineSystemDiskEncryption: testSystemDiskEncryptionConfig,
			MachineFeatures: &v1alpha1.FeaturesConfig{
				RBAC: &testTrue,
			},
			MachineUdev: &v1alpha1.UdevConfig{
				UdevRules: []string{
					"RULE",
				},
			},
			MachineLogging: &v1alpha1.LoggingConfig{
				LoggingDestinations: []v1alpha1.LoggingDestination{
					{
						LoggingEndpoint: &v1alpha1.Endpoint{
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
						},
						LoggingFormat: "json_lines",
					},
				},
			},
			MachineKernel: &v1alpha1.KernelConfig{
				KernelModules: []*v1alpha1.KernelModuleConfig{
					{
						ModuleName: "nvidia",
					},
				},
			},
		},
		ClusterConfig: &v1alpha1.ClusterConfig{
			ClusterID:     testClusterID,
			ClusterSecret: testClusterSecret,
			ControlPlane: &v1alpha1.ControlPlaneConfig{
				Endpoint:           testControlPlaneEndpoint,
				LocalAPIServerPort: 6443,
			},
			ClusterName: testClusterName,
			ClusterNetwork: &v1alpha1.ClusterNetworkConfig{
				CNI:           testCNIConfig,
				DNSDomain:     testServiceDomain,
				PodSubnet:     testPodNet,
				ServiceSubnet: testServiceNet,
			},
			BootstrapToken:                testBundle.Secrets.BootstrapToken,
			ClusterAESCBCEncryptionSecret: testBundle.Secrets.AESCBCEncryptionSecret,
			ClusterCA:                     testBundle.Certs.K8s,
			ClusterAggregatorCA:           testBundle.Certs.K8sAggregator,
			ClusterServiceAccount:         testBundle.Certs.K8sServiceAccount,
			APIServerConfig: &v1alpha1.APIServerConfig{
				ContainerImage: "k8s.gcr.io/kube-apiserver:v1.23.6",
				ExtraArgsConfig: map[string]string{
					"arg": "value",
				},
				ExtraVolumesConfig: []v1alpha1.VolumeMountConfig{
					{
						VolumeHostPath:  "/host",
						VolumeMountPath: "/mount",
						VolumeReadOnly:  true,
					},
				},
				EnvConfig: map[string]string{
					"key": "value",
				},
				CertSANs:                       testSANs,
				DisablePodSecurityPolicyConfig: false,
				AdmissionControlConfig: []*v1alpha1.AdmissionPluginConfig{
					{
						PluginName: "test-plugin",
						PluginConfiguration: v1alpha1.Unstructured{
							Object: map[string]interface{}{
								"key": "value",
							},
						},
					},
				},
			},
			ControllerManagerConfig: &v1alpha1.ControllerManagerConfig{
				ContainerImage: "k8s.gcr.io/kube-controller-manager:v1.23.6",
				ExtraArgsConfig: map[string]string{
					"arg": "value",
				},
				ExtraVolumesConfig: []v1alpha1.VolumeMountConfig{
					{
						VolumeHostPath:  "/",
						VolumeMountPath: "/",
						VolumeReadOnly:  true,
					},
				},
				EnvConfig: map[string]string{
					"env": "value",
				},
			},
			ProxyConfig: &v1alpha1.ProxyConfig{
				Disabled:       false,
				ContainerImage: "k8s.gcr.io/kube-proxy:v1.23.6",
				ModeConfig:     "ipvs",
				ExtraArgsConfig: map[string]string{
					"arg": "value",
				},
			},
			SchedulerConfig: &v1alpha1.SchedulerConfig{
				ContainerImage: "k8s.gcr.io/kube-scheduler:v1.23.6",
				ExtraArgsConfig: map[string]string{
					"arg": "value",
				},
				ExtraVolumesConfig: []v1alpha1.VolumeMountConfig{
					{
						VolumeHostPath:  "/",
						VolumeMountPath: "/",
						VolumeReadOnly:  true,
					},
				},
				EnvConfig: map[string]string{
					"env": "value",
				},
			},
			ClusterDiscoveryConfig: v1alpha1.ClusterDiscoveryConfig{
				DiscoveryEnabled: testDiscoveryEnabled,
				DiscoveryRegistries: v1alpha1.DiscoveryRegistriesConfig{
					RegistryKubernetes: v1alpha1.RegistryKubernetesConfig{
						RegistryDisabled: true,
					},
					RegistryService: v1alpha1.RegistryServiceConfig{
						RegistryDisabled: true,
						RegistryEndpoint: constants.DefaultDiscoveryServiceEndpoint,
					},
				},
			},
			EtcdConfig: &v1alpha1.EtcdConfig{
				ContainerImage: (&v1alpha1.EtcdConfig{}).Image(),
				RootCA:         testBundle.Certs.Etcd,
				EtcdExtraArgs: map[string]string{
					"env": "value",
				},
				EtcdSubnet: "0.0.0.0",
			},
			CoreDNSConfig: &v1alpha1.CoreDNS{
				CoreDNSDisabled: true,
				CoreDNSImage:    (&v1alpha1.CoreDNS{}).Image(),
			},
			ExternalCloudProviderConfig: &v1alpha1.ExternalCloudProviderConfig{
				ExternalEnabled: true,
				ExternalManifests: []string{
					"https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/rbac.yaml",
					"https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/aws-cloud-controller-manager-daemonset.yaml",
				},
			},
			ExtraManifests: []string{
				"https://manifest.org/test.yaml",
			},
			ExtraManifestHeaders: map[string]string{
				"Token":       "1234567",
				"X-ExtraInfo": "info",
			},
			ClusterInlineManifests: v1alpha1.ClusterInlineManifests{
				v1alpha1.ClusterInlineManifest{
					InlineManifestName:     "test-manifest",
					InlineManifestContents: "key: value",
				},
			},
			AdminKubeconfigConfig: &v1alpha1.AdminKubeconfigConfig{
				AdminKubeconfigCertLifetime: time.Hour * 1,
			},
			AllowSchedulingOnMasters: testAllowSchedulingOnMasters,
		},
	}

	nodeData *talosControlNodeResourceData = &talosControlNodeResourceData{
		Name:        wraps("test-node"),
		InstallDisk: wraps("/dev/sda"),
		KernelArgs:  sl("test"),
		CertSANS:    sl("0.0.0.0"),
		ControlPlane: &ControlPlaneConfig{
			Endpoint:           s("https://test"),
			LocalAPIServerPort: types.Int64{Value: 6443},
		},
		Kubelet: &KubeletConfig{
			Image:      wraps("ghcr.io/siderolabs/kubelet:v1.23.6"),
			ClusterDNS: sl("0.0.0.0"),
			ExtraArgs: map[string]types.String{
				"arg": s("value"),
			},
			ExtraMounts: []ExtraMount{
				{
					Destination: wraps("/"),
					Type:        wraps("bind"),
					Source:      wraps("/"),
					Options:     sl("noatime"),
				},
			},
			ExtraConfig:        wraps("key: value"),
			NodeIPValidSubnets: sl("0.0.0.0/0"),
		},
		Pod: sl("key: value"),
		NetworkDevices: map[string]NetworkDevice{
			"wg0": {
				Addresses: sl("0.0.0.0"),
				Wireguard: &Wireguard{
					Peers: []WireguardPeer{
						{
							AllowedIPs:                  sl("0.0.0.0/0"),
							Endpoint:                    wraps("0.0.0.0:0"),
							PersistentKeepaliveInterval: wrapi(25),
							PublicKey:                   wraps("JbHCJXTOS6wRDjZM1an5YHxGz4QsU7VZKim5EBtpMxk="),
						},
					},
					PublicKey:  wraps("JbHCJXTOS6wRDjZM1an5YHxGz4QsU7VZKim5EBtpMxk="),
					PrivateKey: wraps("QDXghbrZoJ+NvLo9MOsaP4JARcPqa0Gy5lXV9EgtNWk="),
				},
			},
			"eth0": {
				Addresses: sl("0.0.0.0"),
				Routes: []Route{
					{
						Network: wraps("0.0.0.0/24"),
						Gateway: wraps("0.0.0.0"),
						Source:  wraps("0.0.0.0"),
						Metric:  wraps("0.0.0.0"),
					},
				},
			},
			"eth1": {},
			"eth2": {},
			"eth3": {},
			"eth4": {},
			"bond0": {
				BondData: &BondData{
					Interfaces:      sl("eth1", "eth2"),
					Mode:            wraps("active-backup"),
					Primary:         wraps("eth1"),
					PrimaryReselect: wraps("better"),
					ArpValidate:     wraps("all"),
				},
			},
			"bond1": {
				BondData: &BondData{
					Interfaces: sl("eth3", "eth4"),
					// Unsupported
					//ARPIPTarget:    sl("0.0.0.0"),
					Mode:           wraps("802.3ad"),
					XmitHashPolicy: wraps("layer2+3"),
					LacpRate:       wraps("fast"),

					// Unsupported
					//AdActorSystem:  wraps("00:00:00:00:00:

					ArpAllTargets: wraps("any"),

					FailoverMac:     wraps("active"),
					AdSelect:        wraps("stable"),
					MiiMon:          wrapi(100),
					UpDelay:         wrapi(200),
					DownDelay:       wrapi(200),
					ArpInterval:     wrapi(0),
					ResendIgmp:      wrapi(0),
					MinLinks:        wrapi(0),
					LpInterval:      wrapi(0),
					PacketsPerSlave: wrapi(0),
					NumPeerNotif:    wrapi(0),
					TlbDynamicLb:    wrapi(0),
					AllSlavesActive: wrapi(0),
					UseCarrier:      wrapb(false),
					AdActorSysPrio:  wrapi(0),
					AdUserPortKey:   wrapi(0),
					PeerNotifyDelay: wrapi(0),
				},
				VLANs: []VLAN{
					{
						Addresses: sl("0.0.0.0"),
						Routes: []Route{
							{
								Network: wraps("0.0.0.0/24"),
								Gateway: wraps("0.0.0.0"),
								Source:  wraps("0.0.0.0"),
								Metric:  wraps("0.0.0.0"),
							},
						},
						DHCP:   wrapb(false),
						VLANId: wrapi(1),
						MTU:    wrapi(1500),
						VIP: &VIP{
							IP:                   wraps("0.0.0.0"),
							EquinixMetalAPIToken: wraps("token"),
							HetznerCloudAPIToken: wraps("token"),
						},
					},
				},
				MTU:  wrapi(1500),
				DHCP: wrapb(false),
				DHCPOptions: &DHCPOptions{
					RouteMetric: wrapi(200),
					IPV4:        wrapb(true),
					IPV6:        wrapb(true),
				},
				Ignore: wrapb(false),
				Dummy:  wrapb(false),
				VIP: &VIP{
					IP:                   wraps("0.0.0.0"),
					EquinixMetalAPIToken: wraps("token"),
					HetznerCloudAPIToken: wraps("token"),
				},
			},
		},
		Nameservers: sl("0.0.0.0"),
		ExtraHost:   map[string][]types.String{"domain": sl("0.0.0.0")},
		Files: []File{
			{
				Content:     wraps("file"),
				Permissions: wrapi(420),
				Path:        wraps("/path"),
				Op:          wraps("create"),
			},
		},
		Env:     map[string]types.String{"var": wraps("value")},
		Sysctls: map[string]types.String{"key": wraps("value")},
		Sysfs:   map[string]types.String{"key": wraps("value")},
		Encryption: &EncryptionData{
			State: &EncryptionConfigData{
				Provider: wraps(testSystemDiskEncryptionConfig.StatePartition.EncryptionProvider),
				Keys: []KeyConfig{
					{
						NodeID:    types.Bool{Null: true},
						KeyStatic: wraps(testSystemDiskEncryptionConfig.StatePartition.EncryptionKeys[0].KeyStatic.KeyData),
						Slot:      wrapi(testSystemDiskEncryptionConfig.StatePartition.EncryptionKeys[0].KeySlot),
					},
				},
				Cipher:      wraps(testSystemDiskEncryptionConfig.StatePartition.EncryptionCipher),
				KeySize:     wrapi(int(testSystemDiskEncryptionConfig.StatePartition.EncryptionKeySize)),
				BlockSize:   wrapi(int(testSystemDiskEncryptionConfig.StatePartition.EncryptionBlockSize)),
				PerfOptions: sl(testSystemDiskEncryptionConfig.StatePartition.EncryptionPerfOptions...),
			},
			Ephemeral: &EncryptionConfigData{
				Provider: wraps(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionProvider),
				Keys: []KeyConfig{
					{
						KeyStatic: types.String{Null: true},
						NodeID:    wrapb(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionKeys[0].KeyNodeID != nil),
						Slot:      wrapi(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionKeys[0].KeySlot),
					},
				},
				Cipher:      wraps(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionCipher),
				KeySize:     wrapi(int(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionKeySize)),
				BlockSize:   wrapi(int(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionBlockSize)),
				PerfOptions: sl(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionPerfOptions...),
			},
		},
		Registry: &Registry{
			Configs: map[string]RegistryConfig{
				"test.org": {
					Username:           wraps("username"),
					Password:           wraps("password"),
					Auth:               wraps("auth"),
					IdentityToken:      wraps("token"),
					ClientCRT:          wraps("test"),
					ClientKey:          wraps("test"),
					CA:                 wraps("test"),
					InsecureSkipVerify: types.Bool{Value: false},
				},
			},
			Mirrors: map[string][]types.String{
				"test.io": sl("test.org"),
			},
		},
		Udev: sl("RULE"),
		MachineControlPlane: &MachineControlPlane{
			ControllerManagerDisabled: types.Bool{Value: false},
			SchedulerDisabled:         types.Bool{Value: false},
		},
		APIServer: &APIServerConfig{
			Image: s("k8s.gcr.io/kube-apiserver:v1.23.6"),
			ExtraArgs: map[string]types.String{
				"arg": s("value"),
			},
			ExtraVolumes: []VolumeMount{
				{
					HostPath:  s("/host"),
					MountPath: s("/mount"),
					Readonly:  types.Bool{Value: true},
				},
			},
			Env: map[string]types.String{
				"key": s("value"),
			},
			CertSANS:   sl("hostname"),
			DisablePSP: types.Bool{Value: false},
			AdmissionPlugins: []AdmissionPluginConfig{
				{
					Name:          s("test-plugin"),
					Configuration: s("key: value"),
				},
			},
		},
		Proxy: &ProxyConfig{
			Image:    s("k8s.gcr.io/kube-proxy:v1.23.6"),
			Mode:     s("ipvs"),
			Disabled: types.Bool{Value: false},
			ExtraArgs: map[string]types.String{
				"arg": s("value"),
			},
		},
		ExtraManifests: sl("https://manifest.org/test.yaml"),
		InlineManifests: []InlineManifest{
			{
				Name:    s("test-manifest"),
				Content: s("key: value"),
			},
		},
		AllowSchedulingOnMasters: types.Bool{Value: false},
	}
)

type runtimeMode struct {
	requiresInstall bool
}

func (m runtimeMode) String() string {
	return fmt.Sprintf("runtimeMode(%v)", m.requiresInstall)
}

func (m runtimeMode) RequiresInstall() bool {
	return m.requiresInstall
}

// TestValidateConfig checks whether an expected valid configuration using values in all fields can be created from a Terraform state struct.
func TestValidateConfig(t *testing.T) {
	genopts := []genv1alpha1.GenOption{genv1alpha1.WithVersionContract(testVersionContract)}

	clock.SetFixedTimestamp(date)

	secrets := &testBundle
	input, err := genv1alpha1.NewInput("test", "https://10.0.1.5", constants.DefaultKubernetesVersion, secrets, genopts...)
	if err != nil {
		t.Fatal(err)
	}

	confString, err := genConfig(machine.TypeControlPlane, input, &nodeData)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := configloader.NewFromBytes([]byte(confString))
	if err != nil {
		t.Fatal(err)
	}

	opts := []config.ValidationOption{config.WithLocal()}
	opts = append(opts, config.WithStrict())

	warnings, err := cfg.Validate(runtimeMode{requiresInstall: true}, opts...)
	for _, w := range warnings {
		t.Logf("%s", w)
	}
	if err != nil {
		t.Fatal(err)
	}
}

// TestConfigDataAll checks if converting a Terraform data struct describing a controlplane node is
// converted into an expected Talos v1alpha1.Config struct.
func TestConfigDataAll(t *testing.T) {
	genopts := []genv1alpha1.GenOption{genv1alpha1.WithVersionContract(testVersionContract)}

	clock.SetFixedTimestamp(date)

	secrets := &testBundle
	input, err := genv1alpha1.NewInput("test", "https://10.0.1.5", constants.DefaultKubernetesVersion, secrets, genopts...)
	if err != nil {
		t.Fatal(err)
	}

	confString, err := genConfig(machine.TypeControlPlane, input, &nodeData)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := configloader.NewFromBytes([]byte(confString))
	if err != nil {
		t.Fatal(err)
	}

	// Build v1alpha1.Config struct from our cfg
	config := v1alpha1.Config{
		ConfigVersion: "v1alpha1",
		ConfigDebug:   false,
		ConfigPersist: true,
		MachineConfig: any(cfg.Machine()).(*v1alpha1.MachineConfig),
		ClusterConfig: any(cfg.Cluster()).(*v1alpha1.ClusterConfig),
	}

	// Struct slice ordering does not matter in this case. so deepequal will fail on equal structures.
	// The diff package from r3labs will be used to provide the same functionality and some
	// useful debugging information as well.

	changes, err := sdiff.Diff(config, *expectedNode)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) > 0 {
		// json marshalindent does not support printing matchers. Ignore it in printout.
		if config.MachineConfig.MachineInstall.InstallDiskSelector != nil {
			config.MachineConfig.MachineInstall.InstallDiskSelector = nil
		}
		expectedNode.MachineConfig.MachineInstall.InstallDiskSelector = nil
		changes, err := sdiff.Diff(config, *expectedNode)
		if err != nil {
			t.Fatal(err)
		}
		changelog, err := json.MarshalIndent(changes, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("expected and actual v1alpha1.Config structs did not match\nchangelog %s", changelog)

		if config.MachineConfig.MachineInstall.InstallDiskSelector == nil {
			t.FailNow()
		}
		matcherChanges, err := sdiff.Diff(
			expectedNode.MachineConfig.MachineInstall.InstallDiskSelector.Size.Matcher,
			config.MachineConfig.MachineInstall.InstallDiskSelector.Size.Matcher,
		)
		if err != nil {
			t.Fatal(err)
		}
		changelog = []byte(spew.Sdump(matcherChanges))
		t.Fatalf("expected and actual matchers did not match\nchangelog %s", changelog)
	}
}

// TestReadControlConfig checks whether we can successfully read a talos Config struct into a Terraform state struct.
func TestReadControlConfig(t *testing.T) {
	var state talosControlNodeResourceData

	state.ReadInto(expectedNode)
	changes, err := sdiff.Diff(state, *nodeData)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) > 0 {
		changelog, err := json.MarshalIndent(changes, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("expected and actual state did not match\nchangelog %s", changelog)
	}
}

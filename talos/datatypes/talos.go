package datatypes

import (
	"time"

	"github.com/talos-systems/crypto/x509"
	"github.com/talos-systems/go-blockdevice/blockdevice/util/disk"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/talos-systems/talos/pkg/machinery/constants"
)

// This file contains methods pertaining to moving data to and from Talos data types.
// It also includes examples for use in unit testing.

// All secrets included can be used to make a valid cluster. For your own sake please
// do not copy them for use in your own cluster.

// Examples section.
var (
	// pointers to true and false. Used in the examples below.
	testFalse bool = false
	testTrue  bool = true

	InputBundleExample = generate.Input{
		Certs:                     SecretsBundleExample.Certs,
		VersionContract:           config.TalosVersionCurrent,
		ControlPlaneEndpoint:      EndpointExample.String(),
		AdditionalSubjectAltNames: apiServerSANsExample,
		AdditionalMachineCertSANs: MachineCertSANsExample,
		ClusterID:                 clusterIDExample,
		ClusterName:               ClusterNameExample,
		ClusterSecret:             clusterSecretExample,
		PodNet:                    podSubnetExample,
		ServiceDomain:             dnsDomainExample,
		ServiceNet:                serviceSubnetExample,
		KubernetesVersion:         constants.DefaultKubernetesVersion,
		Secrets:                   secretsExample,
		TrustdInfo:                trustdInfoExample,
		ExternalEtcd:              false,
		InstallDisk:               installDiskExample,
		InstallImage:              installImageExample,
		InstallExtraKernelArgs:    installKernelArgsExample,
		NetworkConfigOptions:      generate.DefaultGenOptions().NetworkConfigOptions,
		CNIConfig:                 cniConfigExample,
		RegistryMirrors:           machineConfigExample.MachineRegistries.RegistryMirrors,
		RegistryConfig:            machineConfigExample.MachineRegistries.RegistryConfig,
		MachineDisks: []*v1alpha1.MachineDisk{
			machineDiskExample,
		},
		SystemDiskEncryptionConfig: systemDiskEncryptionConfigExample,
		Sysctls:                    sysctlsExample,
		Debug:                      ConfigDebugExample,
		Persist:                    ConfigPersistExample,
		AllowSchedulingOnMasters:   AllowSchedulingOnMastersExample,
		DiscoveryEnabled:           discoveryExample,
	}

	SecretsBundleExample = generate.SecretsBundle{
		Clock:      generate.NewClock(),
		Cluster:    clusterExample,
		Secrets:    secretsExample,
		TrustdInfo: trustdInfoExample,
		Certs:      certsExample,
	}

	clusterExample = &generate.Cluster{
		ID:     "tSuqMd_jk2CU_wGDuPpE7A3HlY9_mcoXWWJ0kRbK8aE=",
		Secret: "6SxVdcxHbbUdSsPpgnnRSHClbxkwmVpxNnbIKVGVirk=",
	}

	secretsExample = &generate.Secrets{
		BootstrapToken:         "3pnxiu.jlzxqodjfujhyado",
		AESCBCEncryptionSecret: "64jlGeK1z13pY0NXKAo7VQHdzJRaugTdTZMflIZErTU=",
	}

	trustdInfoExample = &generate.TrustdInfo{
		Token: "s1pygp.a474wnneqo4v3lbs",
	}

	// TODO investigate whether these values can be truncated and still create a "valid"
	// configuration for unit testing purposes.
	adminCertsExample = &x509.PEMEncodedCertificateAndKey{
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
`)}

	etcdCertsExample = &x509.PEMEncodedCertificateAndKey{
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
	}

	k8sCertsExample = &x509.PEMEncodedCertificateAndKey{
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
	}

	k8sAggregatorExample = &x509.PEMEncodedCertificateAndKey{
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
	}

	k8sServiceAccountExample = &x509.PEMEncodedKey{
		Key: []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIE2bYcsx3UqRIVr8F776i3X44PFfjNq3w5s4OxRvgA+doAoGCCqGSM49
AwEHoUQDQgAEjKovjKS75ObNPsyb2Ury9aP/dXZ9QHwereeXInWAlxzd3ctgDHQN
kGQ1kf6AOXlcAyBOb+KK0LnIh06QCUiZVg==
-----END EC PRIVATE KEY-----
`),
	}

	osExample = &x509.PEMEncodedCertificateAndKey{
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
`)}

	certsExample = &generate.Certs{
		Admin:             adminCertsExample,
		Etcd:              etcdCertsExample,
		K8s:               k8sCertsExample,
		K8sServiceAccount: k8sServiceAccountExample,
		K8sAggregator:     k8sAggregatorExample,
		OS:                osExample,
	}

	MachineConfigExample = &v1alpha1.Config{
		ConfigVersion: "v1alpha1",
		ConfigDebug:   ConfigDebugExample,
		ConfigPersist: ConfigPersistExample,
		MachineConfig: machineConfigExample,
		ClusterConfig: clusterConfigExample,
	}

	clusterConfigExample = &v1alpha1.ClusterConfig{
		ClusterID:                     clusterIDExample,
		ClusterSecret:                 clusterSecretExample,
		ControlPlane:                  controlPlaneConfigExample,
		ClusterName:                   ClusterNameExample,
		ClusterNetwork:                clusterNetworkExample,
		BootstrapToken:                secretsExample.BootstrapToken,
		ClusterAESCBCEncryptionSecret: secretsExample.AESCBCEncryptionSecret,
		ClusterCA:                     k8sCertsExample,
		ClusterAggregatorCA:           k8sAggregatorExample,
		ClusterServiceAccount:         k8sServiceAccountExample,
		APIServerConfig:               apiServerConfigExample,
		ControllerManagerConfig:       controllerManagerExample,
		ProxyConfig:                   proxyConfigExample,
		SchedulerConfig:               schedulerExample,
		ClusterDiscoveryConfig:        clusterDiscoveryExample,
		EtcdConfig:                    etcdConfigExample,
		CoreDNSConfig:                 coreDNSExample,
		ExternalCloudProviderConfig:   externalCloudProviderExample,
		ExtraManifests:                ExtraManifestExample,
		ExtraManifestHeaders:          extraManifestHeaderExample,
		ClusterInlineManifests:        inlineManifestsExample,
		AdminKubeconfigConfig:         adminKubeconfigExample,
		AllowSchedulingOnMasters:      AllowSchedulingOnMastersExample,
	}

	machineConfigExample = &v1alpha1.MachineConfig{
		MachineType:         MachineTypeExample,
		MachineToken:        MachineTokenExample,
		MachineCA:           SecretsBundleExample.Certs.OS,
		MachineCertSANs:     MachineCertSANsExample,
		MachineControlPlane: machineControlPlaneExample,
		MachineKubelet:      machineKubeletExample,
		MachinePods: []v1alpha1.Unstructured{
			{
				Object: machinePodsObjectExample,
			},
		},
		MachineNetwork: networkConfigExample,
		MachineDisks: []*v1alpha1.MachineDisk{
			machineDiskExample,
		},
		MachineInstall: InstallConfigExample,
		MachineFiles: []*v1alpha1.MachineFile{
			machineFileExample,
		},
		MachineEnv: machineEnvExample,
		MachineTime: &v1alpha1.TimeConfig{
			TimeDisabled:    timedisabledExample,
			TimeServers:     timeserversExample,
			TimeBootTimeout: time.Minute * time.Duration(timeoutExample),
		},
		MachineSysctls: sysctlsExample,
		MachineSysfs:   sysfsExample,
		MachineRegistries: v1alpha1.RegistriesConfig{
			RegistryMirrors: map[string]*v1alpha1.RegistryMirrorConfig{
				"docker.io": registryMirrorExample,
			},
			RegistryConfig: map[string]*v1alpha1.RegistryConfig{
				"registry.local": registryConfigExample,
			},
		},
		MachineSystemDiskEncryption: systemDiskEncryptionConfigExample,
		MachineFeatures: &v1alpha1.FeaturesConfig{
			RBAC: &testTrue,
		},
		MachineUdev: &v1alpha1.UdevConfig{
			UdevRules: UdevExample,
		},
		MachineLogging: &v1alpha1.LoggingConfig{
			LoggingDestinations: []v1alpha1.LoggingDestination{
				{
					LoggingEndpoint: &v1alpha1.Endpoint{
						URL: loggingEndpointExample,
					},
					LoggingFormat: loggingFormatExample,
				},
			},
		},
		MachineKernel: &v1alpha1.KernelConfig{
			KernelModules: []*v1alpha1.KernelModuleConfig{
				{
					ModuleName: kernelModuleNameExample,
				},
			},
		},
	}

	InstallConfigExample = &v1alpha1.InstallConfig{
		InstallDisk:            installDiskExample,
		InstallDiskSelector:    installDiskSelectorExample,
		InstallExtraKernelArgs: installKernelArgsExample,
		InstallImage:           installImageExample,
		InstallExtensions: []v1alpha1.InstallExtensionConfig{
			extensionExample,
		},
		InstallBootloader:        installBootloaderExample,
		InstallWipe:              installWipeExample,
		InstallLegacyBIOSSupport: installBiosExample,
	}

	installDiskSelectorExample = &v1alpha1.InstallDiskSelector{
		Size: &v1alpha1.InstallDiskSizeMatcher{
			Matcher: disk.WithName(installMatcherNameExample),
		},
		Name:     diskNameExample,
		Model:    diskModelExample,
		Serial:   diskSerialExample,
		Modalias: diskModaliasExample,
		UUID:     diskUUIDExample,
		WWID:     diskWWIDExample,
		Type:     v1alpha1.InstallDiskType(disk.TypeNVMe),
		BusPath:  diskBusPathExample,
	}

	extensionExample = v1alpha1.InstallExtensionConfig{
		ExtensionImage: extensionImageExample,
	}

	machineFileExample = &v1alpha1.MachineFile{
		FileContent:     machineFileContentExample,
		FilePermissions: v1alpha1.FileMode(machineFilePermissionsExample),
		FilePath:        machineFilePathExample,
		FileOp:          machineFileOpExample,
	}

	machineControlPlaneExample = &v1alpha1.MachineControlPlaneConfig{
		MachineControllerManager: &v1alpha1.MachineControllerManagerConfig{
			MachineControllerManagerDisabled: ControllerManagerDisabledExample,
		},
		MachineScheduler: &v1alpha1.MachineSchedulerConfig{
			MachineSchedulerDisabled: SchedulerDisabledExample,
		},
	}

	machineKubeletExample = &v1alpha1.KubeletConfig{
		KubeletImage:      (&v1alpha1.KubeletConfig{}).Image(),
		KubeletClusterDNS: ClusterDNSExample,
		KubeletExtraArgs:  KubeletExtraArgsExample,
		KubeletExtraMounts: []v1alpha1.ExtraMount{
			kubeletExtraMountExample,
		},
		KubeletExtraConfig: v1alpha1.Unstructured{
			Object: kubeletExtraConfigObjectExample,
		},
		KubeletRegisterWithFQDN: kubeletRegisterWithFQDNExample,
		KubeletNodeIP: v1alpha1.KubeletNodeIPConfig{
			KubeletNodeIPValidSubnets: kubeletSubnetExample,
		},
	}

	kubeletExtraMountExample = v1alpha1.ExtraMount{
		Mount: KubeletMountExample,
	}

	networkConfigExample = &v1alpha1.NetworkConfig{
		NetworkHostname: hostnameExample,
		NetworkInterfaces: []*v1alpha1.Device{
			wireguardExample,
			staticExample,
			dummyExample1,
			dummyExample2,
			dummyExample3,
			dummyExample4,
			activeBackupBondExample,
			lacpBondExample,
			ignoreExample,
			vipExample,
			vlanExample,
			dhcpExample,
		},
		NameServers: nameserversExample,
		ExtraHostEntries: []*v1alpha1.ExtraHost{
			{
				HostIP:      extraHostExampleKey,
				HostAliases: extraHostExampleValues,
			},
		},
		NetworkKubeSpan: v1alpha1.NetworkKubeSpan{
			KubeSpanEnabled: extraHostKubespan,
			// TODO implement in main schemas and data
			KubeSpanAllowDownPeerBypass: false,
		},
	}

	wireguardExample = &v1alpha1.Device{
		DeviceInterface:       "wg0",
		DeviceAddresses:       wgDeviceExampleAddresses,
		DeviceWireguardConfig: wireguardConfigExample,
	}

	wireguardConfigExample = &v1alpha1.DeviceWireguardConfig{
		WireguardPrivateKey:   wgPrivateKeyExample,
		WireguardListenPort:   wgListenPortExample,
		WireguardFirewallMark: wgFirewallMarkExample,
		WireguardPeers: []*v1alpha1.DeviceWireguardPeer{
			{
				WireguardPublicKey:                   wgPublicKeyExample,
				WireguardEndpoint:                    wgEndpointExample,
				WireguardPersistentKeepaliveInterval: time.Duration(wgPersistentKeepaliveExample) * time.Second,
				WireguardAllowedIPs:                  wgAllowedIPsExample,
			},
		},
	}

	staticExample = &v1alpha1.Device{
		DeviceInterface: "eth0",
		DeviceAddresses: staticAddressesExample,
		DeviceRoutes:    routesExample,
		DeviceMTU:       mtuExample,
	}

	dhcpExample = &v1alpha1.Device{
		DeviceInterface: "eth8",
		DeviceDHCP:      true,
	}

	vipExample = &v1alpha1.Device{
		DeviceInterface: "eth7",
		DeviceVIPConfig: vipConfigExample,
	}

	vlanExample = &v1alpha1.Device{
		DeviceInterface: "eth6",
		DeviceVlans:     []*v1alpha1.Vlan{vlanConfigExample},
	}

	ignoreExample = &v1alpha1.Device{
		DeviceInterface: "eth5",
		DeviceIgnore:    true,
	}

	dummyExample1 = &v1alpha1.Device{
		DeviceInterface: "eth1",
		DeviceDummy:     true,
	}

	dummyExample2 = &v1alpha1.Device{
		DeviceInterface: "eth2",
		DeviceDummy:     true,
	}

	dummyExample3 = &v1alpha1.Device{
		DeviceInterface: "eth3",
		DeviceDummy:     true,
	}

	dummyExample4 = &v1alpha1.Device{
		DeviceInterface: "eth4",
		DeviceDummy:     true,
	}

	activeBackupBondExample = &v1alpha1.Device{
		DeviceInterface: "bond0",
		DeviceBond:      activeBackupExample,
	}

	lacpBondExample = &v1alpha1.Device{
		DeviceInterface: "bond1",
		DeviceBond:      lacpExample,
	}

	activeBackupExample = &v1alpha1.Bond{
		BondInterfaces:      abBondInterfaceExample,
		BondMode:            abBondModeExample,
		BondPrimary:         abPrimaryExample,
		BondPrimaryReselect: abPrimaryReselectExample,
		BondARPValidate:     abArpValidateExample,
	}

	lacpExample = &v1alpha1.Bond{
		BondInterfaces:      abLacpInterfacesExample,
		BondMode:            abLacpModeExample,
		BondHashPolicy:      abLacpXmitExample,
		BondLACPRate:        abLacpRateExample,
		BondARPAllTargets:   abLacpArpAllExample,
		BondFailOverMac:     abLacpFailoverMacExample,
		BondADSelect:        abLacpADSelectExample,
		BondMIIMon:          uint32(abLacpMiimonExample),
		BondUpDelay:         uint32(abLacpUpDelayExample),
		BondDownDelay:       uint32(abLacpDownDelayExample),
		BondARPInterval:     uint32(abLacpArpIntervalExample),
		BondResendIGMP:      uint32(abLacpResendIgmpExample),
		BondMinLinks:        uint32(abLacpMinLinksExample),
		BondLPInterval:      uint32(abLacpLPIntervalExample),
		BondPacketsPerSlave: uint32(abLacpBondPacketsPerExample),
		BondNumPeerNotif:    uint8(abLacpNumPeerExample),
		BondTLBDynamicLB:    uint8(abLacpTLBExample),
		BondAllSlavesActive: uint8(abLacpAllSlavesExample),
		BondUseCarrier:      abLacpUseCarrierExample,
		BondADActorSysPrio:  uint16(abLacpAdActorExample),
		BondADUserPortKey:   uint16(abLacpUserPortExample),
		BondPeerNotifyDelay: uint32(abLacpPeerNotifDelayExample),
	}

	vlanConfigExample = &v1alpha1.Vlan{
		VlanAddresses: vlanAddressesExample,
		VlanCIDR:      vlanCIDRExample,
		VlanRoutes:    routesExample,
		VlanDHCP:      vlanDHCPExample,
		VlanID:        uint16(vlanIDExample),
		VlanMTU:       uint32(vlanMTUExample),
		VlanVIP:       vipConfigExample,
	}

	vipConfigExample = &v1alpha1.DeviceVIPConfig{
		SharedIP: vipSharedIPExample,
		EquinixMetalConfig: &v1alpha1.VIPEquinixMetalConfig{
			EquinixMetalAPIToken: vipTokenExample,
		},
		HCloudConfig: &v1alpha1.VIPHCloudConfig{
			HCloudAPIToken: vipTokenExample,
		},
	}

	routesExample = []*v1alpha1.Route{
		{
			RouteNetwork: routeNetworkExample,
			RouteGateway: routeGatewayExample,
		}, {
			RouteNetwork: routeNetworkExample,
			RouteGateway: routeGatewayExample,
			RouteSource:  routeSourceExample,
			RouteMetric:  uint32(routeMetricExample),
		},
	}

	schedulerExample = &v1alpha1.SchedulerConfig{
		ContainerImage:  (&v1alpha1.SchedulerConfig{}).Image(),
		ExtraArgsConfig: schedulerArgsExample,
		ExtraVolumesConfig: []v1alpha1.VolumeMountConfig{
			volumeMountConfigExample,
		},
		EnvConfig: schedulerEnvExample,
	}

	clusterDiscoveryExample = v1alpha1.ClusterDiscoveryConfig{
		DiscoveryEnabled: discoveryExample,
		DiscoveryRegistries: v1alpha1.DiscoveryRegistriesConfig{
			RegistryKubernetes: v1alpha1.RegistryKubernetesConfig{
				RegistryDisabled: discoveryRegistryKubernetesEnabledExample,
			},
			RegistryService: v1alpha1.RegistryServiceConfig{
				RegistryDisabled: discoveryRegistryServiceEnabledExample,
				RegistryEndpoint: constants.DefaultDiscoveryServiceEndpoint,
			},
		},
	}

	etcdConfigExample = &v1alpha1.EtcdConfig{
		ContainerImage: (&v1alpha1.EtcdConfig{}).Image(),
		RootCA:         etcdCertsExample,
		EtcdExtraArgs:  etcdArgsExample,
		EtcdSubnet:     etcdSubnetExample,
	}

	coreDNSExample = &v1alpha1.CoreDNS{
		CoreDNSDisabled: coreDNSDisabledExample,
		CoreDNSImage:    (&v1alpha1.CoreDNS{}).Image(),
	}

	proxyConfigExample = &v1alpha1.ProxyConfig{
		Disabled:        proxyDisabledExample,
		ContainerImage:  (&v1alpha1.ProxyConfig{}).Image(),
		ModeConfig:      proxyModeExample,
		ExtraArgsConfig: proxyArgsExample,
	}

	controllerManagerExample = &v1alpha1.ControllerManagerConfig{
		ContainerImage:  (&v1alpha1.ControllerManagerConfig{}).Image(),
		ExtraArgsConfig: controllerManagerExtraArgsExample,
		ExtraVolumesConfig: []v1alpha1.VolumeMountConfig{
			volumeMountConfigExample,
		},
		EnvConfig: controllerManagerEnvExample,
	}

	apiServerConfigExample = &v1alpha1.APIServerConfig{
		ContainerImage:  (&v1alpha1.APIServerConfig{}).Image(),
		ExtraArgsConfig: apiServerArgsExample,
		ExtraVolumesConfig: []v1alpha1.VolumeMountConfig{
			volumeMountConfigExample,
		},
		EnvConfig:                      apiServerEnvExample,
		CertSANs:                       apiServerSANsExample,
		DisablePodSecurityPolicyConfig: apiServerDisablePSPExample,
		AdmissionControlConfig: []*v1alpha1.AdmissionPluginConfig{
			admissionPluginExample,
		},
	}

	volumeMountConfigExample = v1alpha1.VolumeMountConfig{
		VolumeHostPath:  hostPathExample,
		VolumeMountPath: mountPathExample,
		VolumeReadOnly:  readOnlyExample,
	}

	admissionPluginExample = &v1alpha1.AdmissionPluginConfig{
		PluginName: pluginNameExample,
		PluginConfiguration: v1alpha1.Unstructured{
			Object: pluginObjectExample,
		},
	}

	externalCloudProviderExample = &v1alpha1.ExternalCloudProviderConfig{
		ExternalEnabled:   true,
		ExternalManifests: externalManifestsExample,
	}

	controlPlaneConfigExample = &v1alpha1.ControlPlaneConfig{
		Endpoint:           &v1alpha1.Endpoint{URL: &EndpointExample},
		LocalAPIServerPort: localApiserverExample,
	}

	adminKubeconfigExample = &v1alpha1.AdminKubeconfigConfig{
		AdminKubeconfigCertLifetime: kubeconfigCertExample,
	}

	inlineManifestsExample = v1alpha1.ClusterInlineManifests{
		v1alpha1.ClusterInlineManifest{
			InlineManifestName:     inlineManifestNameExample,
			InlineManifestContents: inlineManifestContentExample,
		},
	}

	clusterNetworkExample = &v1alpha1.ClusterNetworkConfig{
		CNI:           cniConfigExample,
		DNSDomain:     dnsDomainExample,
		PodSubnet:     podSubnetExample,
		ServiceSubnet: serviceSubnetExample,
	}

	cniConfigExample = &v1alpha1.CNIConfig{
		CNIName: constants.CustomCNI,
		CNIUrls: cniURLsExample,
	}

	registryMirrorExample = &v1alpha1.RegistryMirrorConfig{
		MirrorEndpoints: mirrorendpointsExample,
	}

	registryConfigExample = &v1alpha1.RegistryConfig{
		RegistryTLS: &v1alpha1.RegistryTLSConfig{
			TLSClientIdentity: &x509.PEMEncodedCertificateAndKey{
				Crt: []byte(tlsCrtExample),
				Key: []byte(tlsKeyExample),
			},
			TLSCA:                 []byte(tlsCaExample),
			TLSInsecureSkipVerify: tlsInsecureExample,
		},
		RegistryAuth: &v1alpha1.RegistryAuthConfig{
			RegistryUsername:      usernameExample,
			RegistryPassword:      passwordExample,
			RegistryAuth:          authExample,
			RegistryIdentityToken: idtokenExample,
		},
	}

	systemDiskEncryptionConfigExample = &v1alpha1.SystemDiskEncryptionConfig{
		StatePartition: &v1alpha1.EncryptionConfig{
			EncryptionProvider: providerExample,
			EncryptionKeys: []*v1alpha1.EncryptionKey{
				{
					KeyStatic: &v1alpha1.EncryptionKeyStatic{
						KeyData: keydataExample,
					},
					KeySlot: 0,
				},
			},
			EncryptionCipher:      cipherExample,
			EncryptionKeySize:     uint(keysizeExample),
			EncryptionBlockSize:   uint64(blocksizeExample),
			EncryptionPerfOptions: perfoptsExample,
		},
		EphemeralPartition: &v1alpha1.EncryptionConfig{
			EncryptionProvider: providerExample,
			EncryptionKeys: []*v1alpha1.EncryptionKey{
				{
					KeyNodeID: &v1alpha1.EncryptionKeyNodeID{},
					KeySlot:   0,
				},
			},
			EncryptionCipher:      cipherExample,
			EncryptionKeySize:     uint(keysizeExample),
			EncryptionBlockSize:   uint64(blocksizeExample),
			EncryptionPerfOptions: perfoptsExample,
		},
	}

	machineDiskExample = &v1alpha1.MachineDisk{
		DeviceName:     deviceNameExample,
		DiskPartitions: []*v1alpha1.DiskPartition{diskPartitionExample},
	}

	diskPartitionExample = &v1alpha1.DiskPartition{
		DiskSize:       v1alpha1.DiskSize(diskSizeExampleInt),
		DiskMountPoint: diskMountExample,
	}
)

package datatypes

import (
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/opencontainers/runtime-spec/specs-go"
)

var (
	ConfigPersistExample = true
	ConfigDebugExample   = true
)

var (
	MachineTypeExample               = "controlplane"
	MachineTokenExample              = "s1pygp.a474wnneqo4v3lbs"
	MachineCertSANsExample           = []string{"10.0.1.5"}
	SchedulerDisabledExample         = false
	ControllerManagerDisabledExample = false
	ClusterDNSExample                = []string{"10.5.0.1"}
	KubeletExtraArgsExample          = map[string]string{
		"feature-gates": "serverSideApply=true",
	}
	KubeletMountExample = specs.Mount{
		Destination: "/var/lib/example",
		Type:        "bind",
		Source:      "/var/lib/example",
		Options: []string{
			"bind", "rshared", "rw",
		},
	}
	kubeletExtraConfigObjectExample = map[string]interface{}{
		"serverTLSBootstrap": true,
	}
	kubeletExtraConfigExampleString = `serverTLSBootstrap: true
`
	kubeletRegisterWithFQDNExample = false
	kubeletSubnetExample           = []string{"10.0.0.0/8", "!10.0.0.3/32", "fdc7::/16"}

	machinePodsObjectExample = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "pod",
		"metadata": map[string]interface{}{
			"name": "nginx",
		},
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "nginx",
					"image": "nginx",
				},
			},
		},
	}
	MachinePodsStringExample = strings.TrimSpace(`
apiVersion: v1
kind: pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx
`)

	hostnameExample        = "worker-1"
	nameserversExample     = []string{"9.8.7.6", "8.7.6.5"}
	extraHostExampleKey    = "192.168.1.100"
	extraHostExampleValues = []string{"example", "example.domain.tld"}
	extraHostKubespan      = true
)

var (
	wgDeviceExampleAddresses     = []string{"10.12.0.5"}
	wgPrivateKeyExample          = "QDXghbrZoJ+NvLo9MOsaP4JARcPqa0Gy5lXV9EgtNWk="
	wgPublicKeyExample           = "JbHCJXTOS6wRDjZM1an5YHxGz4QsU7VZKim5EBtpMxk="
	wgAllowedIPsExample          = []string{"10.5.0.0/24"}
	wgEndpointExample            = "example.domain.tld:55555"
	wgPersistentKeepaliveExample = 25
	wgFirewallMarkExample        = 5
	wgListenPortExample          = 55555
)

var (
	vlanAddressesExample = []string{"10.0.6.5"}
	vlanDHCPExample      = false
	vlanIDExample        = 1
	vlanMTUExample       = 1500

	vipCIDRExample  = "10.6.0.5"
	vipTokenExample = "token"
)

var (
	staticAddressesExample = []string{"10.5.0.5/24"}
	mtuExample             = 1500

	abBondInterfaceExample   = []string{"eth1", "eth2"}
	abBondModeExample        = "active-backup"
	abPrimaryExample         = "eth1"
	abPrimaryReselectExample = "better"
	abArpValidateExample     = "all"

	abLacpInterfacesExample     = []string{"eth3", "eth4"}
	abLacpModeExample           = "802.3ad"
	abLacpXmitExample           = "layer2+3"
	abLacpRateExample           = "fast"
	abLacpArpAllExample         = "any"
	abLacpFailoverMacExample    = "active"
	abLacpADSelectExample       = "stable"
	abLacpMiimonExample         = 100
	abLacpUpDelayExample        = 200
	abLacpDownDelayExample      = 200
	abLacpArpIntervalExample    = 64
	abLacpResendIgmpExample     = 64
	abLacpMinLinksExample       = 1
	abLacpBondPacketsPerExample = 64
	abLacpLPIntervalExample     = 64
	abLacpNumPeerExample        = 64
	abLacpTLBExample            = 64
	abLacpAllSlavesExample      = 1
	abLacpUseCarrierExample     = &testFalse
	abLacpAdActorExample        = 64
	abLacpUserPortExample       = 1
	abLacpPeerNotifDelayExample = 64
)

var (
	routeNetworkExample = "0.0.0.0/0"
	routeGatewayExample = "10.5.0.1"
	routeSourceExample  = "10.3.0.1"
	routeMetricExample  = 1024

	altRouteNetworkExample = "10.2.0.0/16"
	altGatewayExample      = "10.2.0.1"
)

var (
	machineDisksExample = []any{}

	machineFileContentExample     = "..."
	machineFilePermissionsExample = 438
	machineFilePathExample        = "/tmp/file.txt"
	machineFileOpExample          = "append"

	hostPathExample  = "/var/lib/example"
	mountPathExample = "/var/lib/example"
	readOnlyExample  = false

	dhcpMetricExample = 1024
	dhcpIpv4Example   = false
	dhcpIpv6Example   = true
)

var (
	ClusterNameExample   = "test"
	clusterSecretExample = "6SxVdcxHbbUdSsPpgnnRSHClbxkwmVpxNnbIKVGVirk="
	clusterIDExample     = "tSuqMd_jk2CU_wGDuPpE7A3HlY9_mcoXWWJ0kRbK8aE="

	proxyDisabledExample = false
	proxyModeExample     = "ipvs"
	proxyArgsExample     = map[string]string{
		"proxy-mode": "iptables",
	}

	schedulerArgsExample = map[string]string{
		"feature-gates": "AllBeta=true",
	}
	schedulerArgsTFExample = map[string]types.String{
		"feature-gates": Wraps("AllBeta=true"),
	}
	schedulerEnvExample = map[string]string{
		"key": "value",
	}
	schedulerEnvTFExample = map[string]types.String{
		"key": Wraps("value"),
	}

	etcdArgsTFExample = map[string]types.String{
		"key": Wraps("value"),
	}
	etcdArgsExample = map[string]string{
		"key": "value",
	}

	etcdSubnetExample = "10.0.0.0/8"

	coreDNSDisabledExample = false

	controllerManagerExtraArgsExample = map[string]string{
		"feature-gates": "ServerSideApply=true",
	}

	controllerManagerExtraArgsTFExample = map[string]types.String{
		"feature-gates": Wraps("ServerSideApply=true"),
	}

	controllerManagerEnvTFExample = map[string]types.String{
		"key": Wraps("value"),
	}

	controllerManagerEnvExample = map[string]string{
		"key": "value",
	}

	discoveryExample                          = true
	discoveryRegistryKubernetesEnabledExample = true
	discoveryRegistryServiceEnabledExample    = true

	ExtraManifestExample       = []string{"https://www.example.com/manifest1.yaml", "https://www.example.com/manifest2.yaml"}
	extraManifestHeaderExample = map[string]string{
		"Token":       "1234567",
		"X-ExtraInfo": "info",
	}
	AllowSchedulingOnMastersExample = true
	ExternalManifestsExample        = []string{
		"https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/rbac.yaml",
		"https://raw.githubusercontent.com/kubernetes/cloud-provider-aws/v1.20.0-alpha.0/manifests/aws-cloud-controller-manager-daemonset.yaml",
	}

	localApiserverExample = 443
	EndpointExample       = url.URL{
		Host:   "1.2.3.4",
		Scheme: "https",
	}

	apiServerArgsExample = map[string]string{
		"feature-gates":                    "ServerSideApply=true",
		"http2-max-streams-per-connection": "32",
	}
	apiServerEnvExample = map[string]string{
		"key": "value",
	}
	apiServerSANsExample       = []string{"1.2.3.4", "4.5.6.7"}
	apiServerDisablePSPExample = false

	pluginNameExample   = "PodSecurity"
	pluginConfigExample = strings.TrimSpace(`
apiVersion: pod-security.admission.config.k8s.io/v1alpha1
kind:       PodSecurityConfiguration
defaults:
  enforce:         baseline
  enforce-version: latest
  audit:           restricted
  audit-version:   latest
  warn:            restricted
  warn-version:    latest
exemptions:
  usernames: {}
  runtimeClasses: {}
  namespaces:
  - kube-system
`)
	pluginObjectExample = map[string]interface{}{
		"apiVersion": "pod-security.admission.config.k8s.io/v1alpha1",
		"kind":       "PodSecurityConfiguration",
		"defaults": map[string]interface{}{
			"enforce":         "baseline",
			"enforce-version": "latest",
			"audit":           "restricted",
			"audit-version":   "latest",
			"warn":            "restricted",
			"warn-version":    "latest",
		},
		"exemptions": map[string]interface{}{
			"usernames":      map[string]interface{}{},
			"runtimeClasses": map[string]interface{}{},
			"namespaces":     []interface{}{"kube-system"},
		},
	}

	kubeconfigCertExample        = time.Hour * 8760
	inlineManifestNameExample    = "namespace-ci"
	inlineManifestContentExample = strings.TrimSpace(`
apiVersion: v1
kind: Namespace
metadata:
	name: ci
`)

	dnsDomainExample = "cluster.local"

	podSubnetExample = []string{
		"10.244.0.0/16",
	}
	serviceSubnetExample = []string{
		"10.96.0.0/12",
	}

	installDiskExample       = "/dev/sda"
	installKernelArgsExample = []string{"console=ttyS1", "panic=10"}
	installImageExample      = "ghcr.io/siderolabs/installer:latest"
	installBootloaderExample = true
	installWipeExample       = true
	installBiosExample       = true

	installMatcherNameExample = ""
	diskNameExample           = ""
	diskModelExample          = ""
	diskSerialExample         = ""
	diskModaliasExample       = ""
	diskUUIDExample           = ""
	diskWWIDExample           = ""
	diskTypeExample           = "nvme"
	diskBusPathExample        = ""

	extensionImageExample = ""

	loggingEndpointExample = &url.URL{
		Scheme: "tcp",
		Host:   "1.2.3.4:12345",
	}
	loggingFormatExample    = "json_lines"
	kernelModuleNameExample = "btrfs"

	timedisabledExample  = false
	timeserversExample   = []string{"time.cloudflare.com"}
	timeoutExample       = 2
	timeoutExampleString = "2m"

	sysctlsExample = map[string]string{
		"kernel.domainname":   "talos.dev",
		"net.ipv4.ip_forward": "0",
	}
	sysfsExample = map[string]string{
		"devices.system.cpu.cpu0.cpufreq.scaling_governor": "performance",
	}
	machineEnvExample = map[string]string{
		"GRPC_GO_LOG_VERBOSITY_LEVEL": "99",
		"GRPC_GO_LOG_SEVERITY_LEVEL":  "info",
		"https_proxy":                 "http://DOMAIN\\USERNAME:PASSWORD@SERVER:PORT/",
	}
	UdevExample = []string{
		"SUBSYSTEM==\"drm\", KERNEL==\"renderD*\", GROUP=\"44\", MODE=\"0660\"",
	}

	cniURLsExample = []string{"https://docs.projectcalico.org/archive/v3.20/manifests/canal.yaml"}

	mirrorendpointsExample = []string{"https://registry.local"}
	tlsCrtExample          = "test"
	tlsKeyExample          = "test"
	tlsCaExample           = "test"
	tlsInsecureExample     = false
	usernameExample        = "username"
	passwordExample        = "password"
	authExample            = "auth"
	idtokenExample         = "token"

	keydataExample   = "password"
	providerExample  = "luks2"
	cipherExample    = "aes-xts-plain64"
	keysizeExample   = 4096
	blocksizeExample = 4096
	perfoptsExample  = []string{
		"same_cpu_crypt",
	}

	deviceNameExample = "/dev/sdb1"

	diskSizeExampleInt = 100000000
	diskSizeExampleStr = "100GiB"
	diskMountExample   = "/var/mnt/extra"
)

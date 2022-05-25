resource "talos_control_node" "single_example" {
  # The node's name.
  name = "single-example"

  # Installation options, does not directly configure the node.
  # MAC address for the node. Will be used to apply the initial configuration.
  macaddr = "42:42:42:42:42:42"

  # Expected network range to find the node's interface identified by `macaddr`.
  dhcp_network_cidr = "192.168.122.0/25"

  # Whether this node will be used to bootstrap Etcd. At most a single node per cluster
  # should have this set to true.
  bootstrap = true

  # Node IP address used by the provider to access and send requests to the Talos API.
  # TODO: implement feature mentioned below.
  # If it isn't set the provider will lookup the address of the interface identified by `macaddr`.
  config_ip = "192.168.122.100"

  # The base config from the node's talos_configuration.
  # Contains shared information and secrets.
  base_config = talos_configuration.single_example.base_config

  # Talos options. These will roughly map to options in official Talos documentation.
  # This configuration is targetting version 1.0.5.
  # https://www.talos.dev/v1.0/reference/configuration



  # Machine configuration
  # Install time configuration.
  install = {
	disk  = "/dev/vdb"
	image = "ghcr.io/siderolabs/installer:latest"
	kernel_args = [
	  "console=ttyS1",
	  "panic=10"
	]
  }

  # Extra certificate subject alternative names for the machine’s certificate.
  cert_cans = [
	"10.0.1.5"
  ]

  # Cluster configuration
  # Configure the node's Kubernetes control plane.
  control_plane = {
	endpoint = "https://1.2.3.4"
	local_api_server_port = 443
  }

  # Configure the node's Kubelet.
  kubelet = {
	cluster_dns = [
	  "10.5.0.1"
	]
	extra_args = {
	  "feature-gates": "serverSideApply=true"
	}
	extra_mount = [
	  {
		Source      = "/var/lib/example"
		Destination = "/var/lib/example"
		Type        = "bind",
		options     = ["bind", "rshared", "rw"]
	  }
	]
	extra_config = <<TOC
serverTLSBootstrap: true
TOC
	register_with_fqdn = false
	node_ip_valid_subnets = [
	  "10.0.0.0/8",
	  "!10.0.0.3/32",
	  "fdc7::/16"
	]
  }

  # Static pod definitions that are ran directly on the Kubelet
  pods = [<<TOC
apiVersion: v1
kind: pod
metadata:
  name: nginx
spec:
  containers:
    name: nginx
    image: nginx
TOC
  ]

  # Network configuration. An example including all options can be found in the "guides" section.
  network = {
	# Machine hostname
	hostname = "single-example"
	devices = [
	  {
		# Device interface.
		name = "eth0"
		addresses = [
		  # Device's address in CIDR form.
		  "192.168.122.100/24"
		]
		routes = [
		  {
			network = "0.0.0.0/0"
			# Device's gateway address.
			gateway = "192.168.122.1"
		  }
		]
	  }
	]
	nameservers = [
	  "192.168.122.1"
	]
  }

  # Configure node disk partitions and mounts.
  disks = [
	{
	  device_name = "/dev/sdb1"
	  partitions = [
		{
		  size        = "100GiB"
		  mount_point = "/var/mnt/extra"
		}
	  ]
	}
  ]

  # Write files to the node's root file system. Node that everything outside of /var/ will cause an
  # error when writing to it.
  files = [
	{
	  content     = "..."
	  # Interger form of file permissions eg: integer 438 is octal 666
	  permissions = 438
	  path        = "/tmp/file.txt"
	  op          = "append"
	}
  ]

  # Set node environment variables. All environment variables are set on PID 1 in addition to every service.
  env = {
	"GRPC_GO_LOG_VERBOSITY_LEVEL": "99",
	"GRPC_GO_LOG_SEVERITY_LEVEL":  "info",
	"https_proxy":                 "http://DOMAIN\\USERNAME:PASSWORD@SERVER:PORT/",
  }

  # Set node sysctls.
  sysctls = {
	"kernel.domainname":   "talos.dev",
	"net.ipv4.ip_forward": "0",
  }

  # Set node sysfs.
  sysfs = {
	"devices.system.cpu.cpu0.cpufreq.scaling_governor": "performance",
  }

  # Enable or configure container registry mirrors.
  registry = {
	mirrors = {
	  "ghcr.io": [
         "https://registry.local",
         "https://docker.io/v2/"
	  ]
	}
	configs = {
	  "registry.local": {
		# Options for mutual TLS and adding another trusted CA.
		client_identity_crt = <<TOC
-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
TOC
		client_identity_key = <<TOC
-----BEGIN ENCRYPTED PRIVATE KEY-----
...
-----END ENCRYPTED PRIVATE KEY-----
TOC
		client_identity_ca = <<TOC
-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
TOC
		insecure_skip_verify = false
	  },
	  "ghcr.io": {
		# Registry authentication options. Roughly corresponds with what would be found in a .docker/config.json file.
		username = "username"
		password = "password"
		auth     = "authtoken"
		identity_token = "idtoken"
	  }
	}
  }

  # Configure node system disk encryption.
  encryption = {
	state = {
	  crypt_provider = "luks2"
	  keys = [
		{
		  key_static = "password"
		  slot = 0
		}
	  ]
	  cipher = "aes-xts-plain64"
	  keysize = 4096
	  blocksize = 4096
	  perf_options = [
		"same_cpu_crypt"
	  ]
	}
	ephemeral = {
	  crypt_provider = "luks2"
	  keys = [
		{
		  node_id = true
		  slot = 0
		}
	  ]
	}
  }

  # Configure node udev rules.
  udev = [
	"SUBSYSTEM==\"drm\", KERNEL==\"renderD*\", GROUP=\"44\", MODE=\"0660\""
  ]



  # Cluster wide configuration.
  # Kubernetes controlplane configuration.
  control_plane_config = {
	controller_manager_disabled = false
	scheduler_disabled = false
  }

  # Configure kube-apiserver.
  apiserver = {
	extra_args = {
	  "feature-gates": "ServerSideApply=true",
	  "http2-max-streams-per-connection": "32"
	}
	extra_volumes = [
	  {
		host_path  = "/var/lib/example"
		mount_path = "/var/lib/example"
		readonly   = false
	  }
	]
	env = {
	  "key": "value"
	}
	cert_sans = [
	  "1.2.3.4",
	  "4.5.6.7"
	]
	disable_pod_security_policy = false
	admission_control = [
	  {
		name = "PodSecurity"
		configuration = <<TOC
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
TOC
	  }
	]
  }

  # Configure kube-proxy.
  proxy = {
	mode = "ipvs"
	disabled = false
	extra_args = {
	  "proxy-mode": "iptables"
	}
  }

  # Extra manifests that will be downloaded and applied as a part of the boostrap process.
  extra_manifests = [
	"https://www.example.com/manifest1.yaml",
	"https://www.example.com/manifest2.yaml"
  ]

  # Manifests that will be applied as a part of the boostrap process.
  inline_manifests = [
	{
	  name = "namespace-ci"
	  content = <<TOC
apiVersion: v1
kind: Namespace
metadata:
	name: ci
TOC
	}
  ]

  # Whether pods should be scheduled to master(control plane) nodes.
  allow_scheduling_on_masters = false
  #

  # Bootstrap Etcd as part of the creation process.

}

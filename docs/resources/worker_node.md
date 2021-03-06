---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "talos_worker_node Resource - terraform-provider-talos"
subcategory: ""
description: |-
  Represents a Talos worker node.
---

# talos_worker_node (Resource)

Represents a Talos worker node.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `base_config` (String, Sensitive)
- `config_ip` (String)
- `devices` (Attributes Map) Describes a Talos network device configuration. The map's key is the interface name. (see [below for nested schema](#nestedatt--devices))
- `dhcp_network_cidr` (String)
- `install_disk` (String)
- `macaddr` (String)
- `name` (String)
- `talos_image` (String)

### Optional

- `cert_sans` (List of String) Extra certificate subject alternative names for the machine’s certificate.
- `control_plane` (Attributes) Represents the control plane configuration options. (see [below for nested schema](#nestedatt--control_plane))
- `env` (Map of String) Allows for the addition of environment variables. All environment variables are set on PID 1 in addition to every service.
- `extra_host` (Map of List of String) Allows the addition of user specified files.
- `files` (Attributes List) Describes a machine's files and it's contents and how it will be written to the node's filesystem. (see [below for nested schema](#nestedatt--files))
- `kernel_args` (List of String)
- `kubelet` (Attributes) Represents the kubelet's config values. (see [below for nested schema](#nestedatt--kubelet))
- `nameservers` (List of String) Used to statically set the nameservers for the machine.
- `pod` (List of String) Used to provide static pod definitions to be run by the kubelet directly bypassing the kube-apiserver.
- `proxy` (Attributes) Represents the kube proxy configuration options. (see [below for nested schema](#nestedatt--proxy))
- `registry` (Attributes) Represents the image pull options. (see [below for nested schema](#nestedatt--registry))
- `sysctls` (Map of String) Used to configure the machine’s sysctls.
- `sysfs` (Map of String) Used to configure the machine’s sysctls.
- `udev` (List of String) Configures the udev system.

### Read-Only

- `id` (String) Identifier hash, derived from the node's name.

<a id="nestedatt--devices"></a>
### Nested Schema for `devices`

Required:

- `addresses` (List of String) A list of IP addresses for the interface.
- `name` (String) Network device's Linux interface name.

Optional:

- `bond` (Attributes) Contains the various options for configuring a bonded interface. (see [below for nested schema](#nestedatt--devices--bond))
- `dhcp` (Boolean) Indicates if DHCP should be used to configure the interface.
- `dhcp_options` (Attributes) Specifies DHCP specific options. (see [below for nested schema](#nestedatt--devices--dhcp_options))
- `dummy` (Boolean) Indicates if the interface is a dummy interface..
- `ignore` (Boolean) Indicates if the interface should be ignored (skips configuration).
- `mtu` (Number) The interface’s MTU. If used in combination with DHCP, this will override any MTU settings returned from DHCP server.
- `routes` (Attributes List) Represents a list of routes. (see [below for nested schema](#nestedatt--devices--routes))
- `vip` (Attributes) Contains settings for configuring a Virtual Shared IP on an interface. (see [below for nested schema](#nestedatt--devices--vip))
- `vlans` (Attributes List) Represents vlan settings for a device. (see [below for nested schema](#nestedatt--devices--vlans))
- `wireguard` (Attributes) Contains settings for configuring Wireguard network interface. (see [below for nested schema](#nestedatt--devices--wireguard))

<a id="nestedatt--devices--bond"></a>
### Nested Schema for `devices.bond`

Required:

- `interfaces` (List of String)
- `mode` (String) A bond option. Please see the official kernel documentation.

Optional:

- `ad_actor_sys_prio` (Number) A bond option. Please see the official kernel documentation. Must be a 16 bit unsigned int.
- `ad_actor_system` (String) A bond option. Please see the official kernel documentation.
- `ad_select` (String) A bond option. Please see the official kernel documentation.
- `ad_user_port_key` (Number) A bond option. Please see the official kernel documentation. Must be a 16 bit unsigned int.
- `all_slaves_active` (Number) A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.
- `arp_all_targets` (String) A bond option. Please see the official kernel documentation.
- `arp_interval` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `arp_ip_target` (List of String) A bond option. Please see the official kernel documentation.
- `arp_validate` (String) A bond option. Please see the official kernel documentation.
- `down_delay` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `failover_mac` (String) A bond option. Please see the official kernel documentation.
- `lacp_rate` (String) A bond option. Please see the official kernel documentation.
- `lp_interval` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `mii_mon` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `min_links` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `num_peer_notif` (Number) A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.
- `packets_per_slave` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `peer_notify_delay` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `primary` (String) A bond option. Please see the official kernel documentation.
- `primary_reselect` (String) A bond option. Please see the official kernel documentation.
- `resend_igmp` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `tlb_dynamic_lb` (Number) A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.
- `up_delay` (Number) A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.
- `use_carrier` (Boolean) A bond option. Please see the official kernel documentation.
- `xmit_hash_policy` (String) A bond option. Please see the official kernel documentation.


<a id="nestedatt--devices--dhcp_options"></a>
### Nested Schema for `devices.dhcp_options`

Required:

- `route_metric` (Number) The priority of all routes received via DHCP. Must be castable to a uint32.

Optional:

- `ipv4` (Boolean) Enables DHCPv4 protocol for the interface.
- `ipv6` (Boolean) Enables DHCPv6 protocol for the interface.


<a id="nestedatt--devices--routes"></a>
### Nested Schema for `devices.routes`

Required:

- `network` (String) The route’s network (destination).

Optional:

- `gateway` (String) The route’s gateway (if empty, creates link scope route).
- `metric` (Number) The optional metric for the route.
- `source` (String) The route’s source address.


<a id="nestedatt--devices--vip"></a>
### Nested Schema for `devices.vip`

Required:

- `ip` (String) Specifies the IP address to be used.

Optional:

- `equinix_metal_api_token` (String) Specifies the Equinix Metal API Token.
- `hetzner_cloud_api_token` (String) Specifies the Hetzner Cloud API Token.


<a id="nestedatt--devices--vlans"></a>
### Nested Schema for `devices.vlans`

Required:

- `addresses` (List of String) A list of IP addresses for the interface.

Optional:

- `dhcp` (Boolean) Indicates if DHCP should be used.
- `mtu` (Number) The VLAN’s MTU. Must be a 32 bit unsigned integer.
- `routes` (Attributes List) Represents a list of routes. (see [below for nested schema](#nestedatt--devices--vlans--routes))
- `vip` (Attributes) Contains settings for configuring a Virtual Shared IP on an interface. (see [below for nested schema](#nestedatt--devices--vlans--vip))
- `vlan_id` (Number) The VLAN’s ID. Must be a 16 bit unsigned integer.

<a id="nestedatt--devices--vlans--routes"></a>
### Nested Schema for `devices.vlans.routes`

Required:

- `network` (String) The route’s network (destination).

Optional:

- `gateway` (String) The route’s gateway (if empty, creates link scope route).
- `metric` (Number) The optional metric for the route.
- `source` (String) The route’s source address.


<a id="nestedatt--devices--vlans--vip"></a>
### Nested Schema for `devices.vlans.vip`

Required:

- `ip` (String) Specifies the IP address to be used.

Optional:

- `equinix_metal_api_token` (String) Specifies the Equinix Metal API Token.
- `hetzner_cloud_api_token` (String) Specifies the Hetzner Cloud API Token.



<a id="nestedatt--devices--wireguard"></a>
### Nested Schema for `devices.wireguard`

Required:

- `peers` (Attributes List) A WireGuard device peer configuration. (see [below for nested schema](#nestedatt--devices--wireguard--peers))

Optional:

- `firewall_mark` (Number) Firewall mark for wireguard packets.
- `listen_port` (Number) Listening port for if this node should be a wireguard server.
- `private_key` (String, Sensitive) Specifies a private key configuration (base64 encoded). If one is not provided it is automatically generated and populated this field

Read-Only:

- `public_key` (String) Automatically derived from the private_key field.

<a id="nestedatt--devices--wireguard--peers"></a>
### Nested Schema for `devices.wireguard.peers`

Required:

- `allowed_ips` (List of String) AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.
- `endpoint` (String) Specifies the endpoint of this peer entry.
- `public_key` (String) Specifies the public key of this peer.

Optional:

- `persistent_keepalive_interval` (Number) Specifies the persistent keepalive interval for this peer. Provided in seconds.




<a id="nestedatt--control_plane"></a>
### Nested Schema for `control_plane`

Optional:

- `endpoint` (String) Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.
- `local_api_server_port` (Number) The port that the API server listens on internally. This may be different than the port portion listed in the endpoint field.


<a id="nestedatt--files"></a>
### Nested Schema for `files`

Required:

- `content` (String) The file's content. Not required to be base64 encoded.
- `op` (String) Mode for the file. Can be one of create, append and overwrite.
- `path` (String) Full path for the file to be created at.
- `permissions` (Number) Unix permission for the file


<a id="nestedatt--kubelet"></a>
### Nested Schema for `kubelet`

Optional:

- `cluster_dns` (List of String) An optional reference to an alternative kubelet clusterDNS ip list.
- `extra_args` (Map of String) Used to provide additional flags to the kubelet.
- `extra_config` (String) The extraConfig field is used to provide kubelet configuration overrides. Must be valid YAML
- `extra_mount` (Attributes List) Wraps the OCI Mount specification. (see [below for nested schema](#nestedatt--kubelet--extra_mount))
- `image` (String) An optional reference to an alternative kubelet image.
- `node_ip_valid_subnets` (List of String) The validSubnets field configures the networks to pick kubelet node IP from.
- `register_with_fqdn` (Boolean) Used to force kubelet to use the node FQDN for registration. This is required in clouds like AWS.

<a id="nestedatt--kubelet--extra_mount"></a>
### Nested Schema for `kubelet.extra_mount`

Required:

- `destination` (String) Destination of mount point: path inside container. This value MUST be an absolute path.
- `source` (String) A device name, but can also be a file or directory name for bind mounts or a dummy. Path values for bind mounts are either absolute or relative to the bundle. A mount is a bind mount if it has either bind or rbind in the options.

Optional:

- `options` (List of String) Mount options of the filesystem to be used.
- `type` (String) The type of the filesystem to be mounted.



<a id="nestedatt--proxy"></a>
### Nested Schema for `proxy`

Optional:

- `extra_args` (Map of String) Extra arguments to supply to kube-proxy.
- `image` (String) The container image used in the kube-proxy manifest.
- `is_disabled` (Boolean) Disable kube-proxy deployment on cluster bootstrap.
- `mode` (String) The container image used in the kube-proxy manifest.


<a id="nestedatt--registry"></a>
### Nested Schema for `registry`

Optional:

- `configs` (Attributes Map) Specifies TLS & auth configuration for HTTPS image registries. The meaning of each auth_field is the same with the corresponding field in .docker/config.json.

Key description: The first segment of an image identifier, with ‘docker.io’ being default one. To catch any registry names not specified explicitly, use ‘*’. (see [below for nested schema](#nestedatt--registry--configs))
- `mirrors` (Map of List of String) Specifies mirror configuration for each registry.

<a id="nestedatt--registry--configs"></a>
### Nested Schema for `registry.configs`

Optional:

- `auth` (String, Sensitive) Auth for optional registry authentication.
- `ca` (String) CA registry certificate to add the list of trusted certificates. Non base64 encoded.
- `client_identity_crt` (String, Sensitive) Enable mutual TLS authentication with the registry. Non base64 encoded client certificate.
- `client_identity_key` (String, Sensitive) Enable mutual TLS authentication with the registry. Non base64 encoded client key.
- `identity_token` (String, Sensitive) Identity token for optional registry authentication.
- `insecure_skip_verify` (Boolean) Skip TLS server certificate verification (not recommended)..
- `password` (String, Sensitive) Password for optional registry authentication.
- `username` (String) Username for optional registry authentication.



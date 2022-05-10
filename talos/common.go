package talos

import (
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opencontainers/runtime-spec/specs-go"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Schema validation helpers

// validateDomain checks whether the provided schema value is a valid domain name.
func validateDomain(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain, got \"%s\"", key, v))
	}
	return
}

// validateDomain checks whether the provided schema value is a valid MAC address.
func validateMAC(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, err := net.ParseMAC(v); err != nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid MAC address, got \"%s\", error \"%s\"", key, v, err.Error()))
	}
	return
}

// validateDomain checks whether the provided schema value is a valid MAC address.
func validateCIDR(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, _, err := net.ParseCIDR(v); err != nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid CIDR IP address, got \"%s\", error \"%s\"", key, v, err.Error()))
	}
	return
}

// validateIP checks whether the provided schema value is a valid IP address.
func validateIP(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if net.ParseIP(v) == nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid IP address, got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid hostname.
func validateHost(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*:[0-9]{2,}`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain with a port appended, seperated by \":\", got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid wireguard endpoint.
func validateEndpoint(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(:[0-9]{2,})?`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain with an optional port appended, seperated by \":\", got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid image registry identifier. Used for specifying images for containers in Static Pods.
func validateImage(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[^/]+\.[^/.]+/([^/.]+/)?[^/.]+(:.+)?`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a valid container image, got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid wireguard public or private key.
func validateKey(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, err := wgtypes.ParseKey(v); err != nil {
		errs = append(errs, err)
	}
	return
}

func StringListSchema(desc string) schema.Schema {
	return schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: desc,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func StringListSchemaValidate(desc string, validate schema.SchemaValidateFunc) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: desc,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validate,
		},
	}
}

// Validate is a helper for building terraform resource schemas that denotes the passed schema is Required.
func Required(in *schema.Schema) (out *schema.Schema) {
	out = in
	out.Required = true
	return
}

// Validate is a helper for building terraform resource schemas that denotes the passed schema is Optional.
func Optional(in *schema.Schema) (out *schema.Schema) {
	out = in
	out.Optional = true
	return
}

// Validate is a helper for building terraform resource schemas that adds validation to the passed schema.
func Validate(in *schema.Schema, validate schema.SchemaValidateFunc) (out *schema.Schema) {
	out = in
	out.ValidateFunc = validate
	return
}

// StringMap is a helper for building terraform resource schemas that adds validation to the elements to the passed schema.
func ValidateInner(in *schema.Schema, validate schema.SchemaValidateFunc) (out *schema.Schema) {
	out = in
	out.Elem.(*schema.Schema).ValidateFunc = validate
	return
}

// StringList is a helper for building terraform resource schemas that creates a basic string TypeList.
func StringList(desc string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		Description: desc,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

// StringMap is a helper for building terraform resource schemas that creates a basic string TypeMap
func StringMap(desc string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Description: desc,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

// Common schemas used by the controlplane and worker node resources
var (
	// VolumeMountSchema Describes extra volume mount for the static pods.
	// See https://www.talos.dev/v1.0/reference/configuration/#volumemountconfig for more information.
	VolumeMountSchema schema.Resource = schema.Resource{
		Schema: map[string]*schema.Schema{
			"host_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path on the host.",
				// TODO validate it is a well formed path
			},
			"mount_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path in the container.",
				// TODO validate it is a well formed path
			},
			"readonly": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Mount the volume read only.",
			},
		},
	}

	// ExtraMountListSchema wraps the OCI mount specification
	// For more information see https://github.com/opencontainers/runtime-spec/blob/main/config.md#mounts
	ExtraMountListSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Wraps the OCI Mount specification.",
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Destination of mount point: path inside container. This value MUST be an absolute path.",
				},
				"type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The type of the filesystem to be mounted.",
					ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
						v := value.(string)
						// Must not contain spaces
						pattern := `\s`
						if regexp.MustCompile(pattern).Match([]byte(v)) {
							errs = append(errs, fmt.Errorf("%s: Invalid mount type, must be a valid platform filesystem type, got \"%s\"\nSee OCI mount specs for more information", key, v))
						}
						return
					},
				},
				"source": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A device name, but can also be a file or directory name for bind mounts or a dummy. Path values for bind mounts are either absolute or relative to the bundle. A mount is a bind mount if it has either bind or rbind in the options.",
					//ValidateFunc: validatePath,
				},
				"options": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Mount options of the filesystem to be used.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
							v := value.(string)
							// Must not contain spaces
							pattern := `\s`
							if regexp.MustCompile(pattern).Match([]byte(v)) {
								errs = append(errs, fmt.Errorf("%s: Invalid mount option, must be a valid platform mount option, got \"%s\"\nSee OCI mount specs for more information", key, v))
							}
							return
						},
					},
				},
			},
		},
	}

	// KubeletConfigSchema represents the kubelet config values.
	// see https://www.talos.dev/v1.0/reference/configuration/#kubeletconfig for more information
	KubeletConfigSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Represents the kubelet config values.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"image": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "An optional reference to an alternative kubelet image.",
					ValidateFunc: validateImage,
				},
				"cluster_dns": StringListSchemaValidate("An optional reference to an alternative kubelet clusterDNS ip list.", validateIP),
				"extra_args": {
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "Used to provide additional flags to the kubelet.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"extra_mount": &ExtraMountListSchema,
				// TODO Add yaml validation function
				"extra_config": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The extraConfig field is used to provide kubelet configuration overrides. Must be valid YAML",
				},
				"register_with_fqdn": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Used to force kubelet to use the node FQDN for registration. This is required in clouds like AWS.",
				},
				"node_ip_valid_subnets": StringListSchemaValidate("The validSubnets field configures the networks to pick kubelet node IP from.", validateCIDR),
			},
		},
	}

	kubeletExtraMountSchema schema.Schema = ExtraMountListSchema

	// RegistryListSchema represents the image pull options.
	// See https://www.talos.dev/v1.0/reference/configuration/#registriesconfig for more information
	RegistryListSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MinItems:    1,
		Description: "Represents the image pull options.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"mirror": {
					Type:        schema.TypeList,
					Optional:    true,
					MinItems:    0,
					Description: "Specifies mirror configuration for each registry.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"registry_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The first segment of image identifier, with ‘docker.io’ being default one. To catch any registry names not specified explicitly, use ‘*’.",
							},
							"endpoints": ValidateInner(Required(StringList("List of endpoints (URLs) for registry mirrors to use.")), validateEndpoint),
						},
					},
				},
				"config": {
					Type:        schema.TypeList,
					Optional:    true,
					MinItems:    0,
					Description: "Specifies TLS & auth configuration for HTTPS image registries. The meaning of each auth_field is the same with the corresponding field in .docker/config.json.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"registry_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The first segment of image identifier, with ‘docker.io’ being default one. To catch any registry names not specified explicitly, use ‘*’.",
							},
							"username": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Username for optional registry authentication.",
							},
							"password": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Password for optional registry authentication.",
							},
							"auth": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Auth for optional registry authentication.",
							},
							"identity_token": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Identity token for optional registry authentication.",
							},
							"client_identity_crt": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Enable mutual TLS authentication with the registry. Base64 encoded client certificate.",
								// TODO: validate it's a correctly encoded PEM certificate
							},
							"client_identity_key": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Enable mutual TLS authentication with the registry. Base64 encoded client key.",
								// TODO: validate it's a correctly encoded PEM key
							},
							"ca": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "CA registry certificate to add the list of trusted certificates. Should be Base64 encoded.",
								// TODO: Verify CA is base64 encoded
							},
							"insecure_skip_verify": {
								Type:        schema.TypeBool,
								Optional:    true,
								Default:     false,
								Description: "Skip TLS server certificate verification (not recommended)..",
							},
						},
					},
				},
			},
		},
	}

	// AddressesSchema describes a list of IP addresses
	AddressesListSchema schema.Schema = *StringListSchemaValidate("The addresses in CIDR notation or as plain IPs to use.", validateCIDR)

	// networkInterfaceSchema describes a Talos Device configuration.
	// For more information refer to https://www.talos.dev/v1.0/reference/configuration/#device
	networkInterfaceSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: "Describes a Talos Device configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The interface name.",
				},
				"addresses": &AddressesListSchema,
				"route":     &RouteListSchema,
				"bond":      &BondSchema,
				"vlan":      &VlanListSchema,
				"mtu": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The interface’s MTU. If used in combination with DHCP, this will override any MTU settings returned from DHCP server.",
				},
				"dhcp": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Indicates if DHCP should be used to configure the interface.",
				},
				"ignore": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Indicates if the interface should be ignored (skips configuration).",
				},
				"dummy": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Indicates if the interface is a dummy interface..",
				},
				"dhcp_options": &DHCPOptionsSchema,
				"wireguard":    &WireguardConfigSchema,
				"vip":          &VipSchema,
			},
		},
	}

	// BondSchema contains the various options for configuring a bonded interface.
	// See https://www.talos.dev/v1.0/reference/configuration/#bond for more information about the Bond configuration options.
	BondSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Contains the various options for configuring a bonded interface.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"mode": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"xmit_hash_policy": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"lacp_rate": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"ad_actor_system": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"arp_validate": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"arp_all_targets": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"primary": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"primary_reselect": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"failover_mac": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"ad_select": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"mii_mon": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"up_delay": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"down_delay": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"arp_interval": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"resend_igmp": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"min_links": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"lp_interval": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"packets_per_slave": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
				"num_peer_notif": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.",
				},
				"tlb_dynamic_lb": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.",
				},
				"all_slaves_active": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 8 bit unsigned int.",
				},
				"use_carrier": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation.",
				},
				"ad_actor_sys_prio": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 16 bit unsigned int.",
				},
				"ad_user_port_key": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 16 bit unsigned int.",
				},
				"peer_notify_delay": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "A bond option. Please see the official kernel documentation. Must be a 32 bit unsigned int.",
				},
			},
		},
	}

	// DHCPOptionsSchema specifies DHCP specific options.
	DHCPOptionsSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Specifies DHCP specific options.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"route_metric": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The priority of all routes received via DHCP. Must be castable to a uint32.",
				},
				"ipv4": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Enables DHCPv4 protocol for the interface.",
				},
				"ipv6": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Enables DHCPv6 protocol for the interface.",
				},
			}},
	}

	// VlanSchema represents vlan settings for a device.
	VlanListSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MinItems:    1,
		Description: "Represents vlan settings for a device.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"addresses": &AddressesListSchema,
				"routes":    &RouteListSchema,
				"dhcp": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Indicates if DHCP should be used.",
				},
				"vlan_id": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The VLAN’s ID. Must be a 16 bit unsigned integer.",
				},
				"mtu": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The VLAN’s MTU. Must be a 32 bit unsigned integer.",
				},
				"vip": &VipSchema,
			},
		},
	}

	// See https://www.talos.dev/v1.0/reference/configuration/#devicevipconfig for more information.
	VipSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Contains settings for configuring a Virtual Shared IP on an interface..",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ip": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateIP,
					Description:  "Specifies the IP address to be used.",
				},
				"equinix_metal_api_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Specifies the Equinix Metal API Token.",
				},
				"h_cloud_api_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Specifies the Hetzner Cloud API Token.",
				},
			},
		},
	}

	// RouteSchema represents a network route.
	// See https://www.talos.dev/v1.0/reference/configuration/#route for more information
	RouteListSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MinItems:    1,
		Description: "Represents a list of routes.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"network": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateCIDR,
					Description:  "The route’s network (destination).",
				},
				"gateway": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateIP,
					Description:  "The route’s gateway (if empty, creates link scope route).",
				},
				"source": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateIP,
					Description:  "The route’s source address.",
				},
				"metric": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The optional metric for the route.",
				},
			},
		},
	}

	// WireguardConfigSchema describes a Talos network interface wireguard configuration
	// for more information refer to https://www.talos.dev/v1.0/reference/configuration/#devicewireguardconfig
	WireguardConfigSchema schema.Schema = schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Contains settings for configuring Wireguard network interface.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"peer": {
					Type:        schema.TypeList,
					Required:    true,
					MinItems:    1,
					Description: "A WireGuard device peer configuration.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"allowed_ips": {
								Type:        schema.TypeList,
								Required:    true,
								MinItems:    1,
								Description: "AllowedIPs specifies a list of allowed IP addresses in CIDR notation for this peer.",
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validateCIDR,
								},
							},
							"endpoint": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validateEndpoint,
								Description:  "Specifies the endpoint of this peer entry.",
							},
							"persistent_keepalive_interval": {
								Type:     schema.TypeInt,
								Optional: true,
								ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
									v := value.(int)
									if v < 0 {
										errs = append(errs, fmt.Errorf("%s: Persistent keepalive interval must be a positive integer, got %d", key, v))
									}
									return
								},
								Description: "Specifies the persistent keepalive interval for this peer. Provided in seconds.",
								Default:     0,
							},
							"public_key": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validateKey,
								Description:  "Specifies the public key of this peer.",
							},
						},
					},
				},
				"public_key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Automatically derived from the private_key field.",
				},
				"private_key": {
					Type:         schema.TypeString,
					Sensitive:    true,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validateKey,
					Description:  "Specifies a private key configuration (base64 encoded). If one is not provided it is automatically generated and populated this field",
				},
			},
		},
	}
)

// Type aliases to ease working with interfaces
type TypeMap = map[string]interface{}
type TypeList = []interface{}

// Valid primitive types that can be inside a Terraform TypeList
type TypeListT interface {
	string | int | bool
}

// Valid primitive types that can be values of a Terraform TypeMap
type TypeMapT interface {
	string | int | bool
}

// Helpers for simple assignment of key/value maps and lists of primitive types

// ExpandTypeList safely extracts values from an interface list (TypeList) that contains primitve types (specified by TypeListT)
// and returns a list of that primitive type. It's used to extract values from TypeList fields of Resource schemas.
func ExpandTypeList[T TypeListT](typelist TypeList) (result []T) {
	result = []T{}
	for _, val := range typelist {
		result = append(result, val.(T))
	}
	return
}

// ExpandTypeMap safely extracts values from a map of string keys to interfaces (TypeMap) that contains primitve types
// (specified by TypeListT) and returns a map of string keys to that primitive type. It's used to extract values from TypeMap fields
// of Resource schemas.
func ExpandTypeMap[T TypeMapT](typelist TypeMap) (result map[string]T) {
	result = map[string]T{}
	for k, val := range typelist {
		result[k] = val.(T)
	}
	return
}

// Errors for the ExpandDeviceList function
var (
	WireguardExtraFieldError = fmt.Errorf("There can only be one wireguard field in each network device.")
)

// ExpandRoutes extracts the values from the "route" schema value used in network interfaces and returns a list of pointers to Talos routes.
func ExpandRoutes(interfaces TypeList) (routes []*v1alpha1.Route) {
	// Expand the list of interface maps that form's the "route" schema
	for _, resourceRoute := range interfaces {
		r := resourceRoute.(map[string]interface{})

		route := &v1alpha1.Route{
			RouteGateway: r["gateway"].(string),
			RouteNetwork: r["network"].(string),
		}

		if r["metric"] != nil {
			route.RouteMetric = uint32(r["metric"].(int))
		}

		routes = append(routes, route)
	}

	return
}

// ExpandBondConfig extracts the values from the "bond" schema used in the network interface schema and returns
// a pointer to a Talos Bond. Despite returning an error it currently has no error conditions, the value
// is there for future use.
func ExpandBondConfig(bondSchema TypeMap) (bond *v1alpha1.Bond, err error) {
	for _, iface := range bondSchema["interfaces"].(TypeList) {
		bond.BondInterfaces = append(bond.BondInterfaces, iface.(string))
	}

	bond.BondMode = bondSchema["mode"].(string)

	if bondSchema["xmit_hash_policy"] != nil {
		bond.BondHashPolicy = bondSchema["xmit_hash_policy"].(string)
	}

	bond.BondLACPRate = bondSchema["lacp_rate"].(string)

	if bondSchema["arp_ip_targets"] != nil {
		for _, arpIpTarget := range bondSchema["arp_ip_targets"].(TypeList) {
			bond.BondARPIPTarget = append(bond.BondARPIPTarget, arpIpTarget.(string))
		}
	}

	if bondSchema["ad_actor_system"] != nil {
		bond.BondADActorSystem = bondSchema["ad_actor_system"].(string)
	}
	if bondSchema["arp_validate"] != nil {
		bond.BondARPValidate = bondSchema["arp_validate"].(string)
	}
	if bondSchema["arp_all_targets"] != nil {
		bond.BondARPAllTargets = bondSchema["arp_all_targets"].(string)
	}
	if bondSchema["primary"] != nil {
		bond.BondPrimary = bondSchema[""].(string)
	}
	if bondSchema["primary_reselect"] != nil {
		bond.BondPrimaryReselect = bondSchema["primary_reselect"].(string)
	}
	if bondSchema["failover_mac"] != nil {
		bond.BondFailOverMac = bondSchema["failover_mac"].(string)
	}
	if bondSchema["ad_select"] != nil {
		bond.BondADSelect = bondSchema["ad_select"].(string)
	}
	if bondSchema["mii_mon"] != nil {
		bond.BondMIIMon = uint32(bondSchema["mii_mon"].(int))
	}
	if bondSchema["up_delay"] != nil {
		bond.BondUpDelay = uint32(bondSchema["up_delay"].(int))
	}
	if bondSchema["down_delay"] != nil {
		bond.BondDownDelay = uint32(bondSchema["down_delay"].(int))
	}
	if bondSchema["arp_interval"] != nil {
		bond.BondARPInterval = uint32(bondSchema["arp_interval"].(int))
	}
	if bondSchema["resend_igmp"] != nil {
		bond.BondResendIGMP = uint32(bondSchema["resend_igmp"].(int))
	}
	if bondSchema["min_links"] != nil {
		bond.BondMinLinks = uint32(bondSchema["min_links"].(int))
	}
	if bondSchema["lp_interval"] != nil {
		bond.BondLPInterval = uint32(bondSchema["lp_interval"].(int))
	}
	if bondSchema["packets_per_slave"] != nil {
		bond.BondPacketsPerSlave = uint32(bondSchema["packets_per_slave"].(int))
	}
	if bondSchema["num_peer_notif"] != nil {
		bond.BondNumPeerNotif = uint8(bondSchema["num_peer_notif"].(int))
	}
	if bondSchema["tlb_dynamic_lb"] != nil {
		bond.BondTLBDynamicLB = uint8(bondSchema["tlb_dynamic_lb"].(int))
	}
	if bondSchema["all_slaves_active"] != nil {
		bond.BondAllSlavesActive = uint8(bondSchema["all_slaves_active"].(int))
	}
	if bondSchema["use_carrier"] != nil {
		*bond.BondUseCarrier = bondSchema["use_carrier"].(bool)
	}
	if bondSchema["ad_actor_sys_prio"] != nil {
		bond.BondADActorSysPrio = uint16(bondSchema["ad_actor_sys_prio"].(int))
	}
	if bondSchema["ad_user_port_key"] != nil {
		bond.BondADUserPortKey = uint16(bondSchema["ad_user_port_key"].(int))
	}
	if bondSchema["peer_notify_delay"] != nil {
		bond.BondPeerNotifyDelay = uint32(bondSchema["peer_notify_delay"].(int))
	}

	return
}

// ExpandWireguardConfig extracts the values from the "wireguard" schema value used in network interfaces and returns
// a pointer to a Talos DeviceWireguardConfig. If a private key is not provided it will generate one using the generator
// from the wgtypes library.
func ExpandWireguardConfig(wireguardConfig TypeMap) (wgConfig *v1alpha1.DeviceWireguardConfig, err error) {
	wg := wireguardConfig

	// Expand the list of interface maps that form's the "wireguard" schema's "peer" element
	peers := []*v1alpha1.DeviceWireguardPeer{}
	for _, peer := range wg["peer"].([]interface{}) {
		p := peer.(map[string]interface{})

		peers = append(peers, &v1alpha1.DeviceWireguardPeer{
			WireguardAllowedIPs:                  ExpandTypeList[string](p["allowed_ips"].(TypeList)),
			WireguardEndpoint:                    p["endpoint"].(string),
			WireguardPersistentKeepaliveInterval: time.Duration(p["persistent_keepalive_interval"].(int)) * time.Second,
			WireguardPublicKey:                   p["public_key"].(string),
		})
	}

	pk := wg["private_key"].(string)
	if pk == "" {
		pk_, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		pk = pk_.String()
	}

	wgConfig = &v1alpha1.DeviceWireguardConfig{
		WireguardPrivateKey: pk,
		WireguardPeers:      peers,
	}

	if wg["firewall_mark"] != nil {
		wgConfig.WireguardFirewallMark = wg["firewall_mark"].(int)
	}

	if wg["listen_port"] != nil {
		wgConfig.WireguardListenPort = wg["listen_port"].(int)
	}

	return
}

// ExpandsDeviceList extracts the values from the "interface" schema value shared between node resources and returns
// a list of pointers to Talos network devices. It needs to handle an optional "wireguard" field, max of one, and optional route fields.
// See https://www.talos.dev/v1.0/reference/configuration/#device for more information about the output spec.
func ExpandDeviceList(interfaces []interface{}) (devices []*v1alpha1.Device, err error) {
	for _, netInterface := range interfaces {
		n := netInterface.(map[string]interface{})
		dev := &v1alpha1.Device{}

		// The interface name
		dev.DeviceInterface = n["name"].(string)
		// Static IP addresses for the interface
		dev.DeviceAddresses = ExpandTypeList[string](n["addresses"].(TypeList))

		if len(n["route"].(TypeList)) > 0 {
			dev.DeviceRoutes = ExpandRoutes(n["route"].(TypeList))
		}

		// The interface's MTU
		if n["mtu"] != nil {
			dev.DeviceMTU = n["mtu"].(int)
		}

		// Does the interface use DHCP for auto configuration?
		if n["dhcp"] != nil {
			dev.DeviceDHCP = n["dhcp"].(bool)
		}

		if n["dhcpOptions"] != nil {
			d := n["dhcpOptions"].(map[string]interface{})
			dhcp := &v1alpha1.DHCPOptions{
				DHCPRouteMetric: uint32(d["route_metric"].(int)),
			}

			if d["ipv4"] != nil {
				*dhcp.DHCPIPv4 = d["ipv4"].(bool)
			}

			if d["ipv6"] != nil {
				*dhcp.DHCPIPv6 = d["ipv6"].(bool)
			}

			dev.DeviceDHCPOptions = dhcp
		}

		wg_ := n["wireguard"].(TypeList)
		// Ensure there can only be one instance of the wireguard interface, Return an error if there are more
		if len(wg_) > 1 {
			return nil, WireguardExtraFieldError
		}
		if len(wg_) == 1 {
			dev.DeviceWireguardConfig, err = ExpandWireguardConfig(wg_[0].(TypeMap))
			if err != nil {
				return nil, err
			}
		}

		if len(n["vip"].(TypeList)) == 1 {
			vipSchema := n["vip"].(TypeList)[0].(TypeMap)
			vip := v1alpha1.DeviceVIPConfig{
				SharedIP: vipSchema["ip"].(string),
			}

			// Would be nice if I could acceptance test these
			if vipSchema["equinix_metal"] != nil {
				equinix := vipSchema["equinix_metal"].(TypeMap)
				vip.EquinixMetalConfig = &v1alpha1.VIPEquinixMetalConfig{
					EquinixMetalAPIToken: equinix["api_token"].(string),
				}
			}

			if vipSchema["hcloud"] != nil {
				hetzner := vipSchema["hcloud"].(TypeMap)
				vip.HCloudConfig = &v1alpha1.VIPHCloudConfig{
					HCloudAPIToken: hetzner["api_token"].(string),
				}
			}

			dev.DeviceVIPConfig = &vip
		}

		if len(n["bond"].(TypeList)) == 1 {
			dev.DeviceBond, err = ExpandBondConfig(n["bond"].(TypeMap))
			if err != nil {
				return nil, err
			}
		}

		if len(n["vlan"].(TypeList)) == 1 {
			for _, vlanSchema := range n["vlan"].(TypeList) {
				vlan := vlanSchema.(TypeMap)
				dev.DeviceVlans = append(dev.DeviceVlans, &v1alpha1.Vlan{
					VlanAddresses: ExpandTypeList[string](vlan["addresses"].(TypeList)),
					VlanRoutes:    ExpandRoutes(vlan["route"].(TypeList)),
					VlanDHCP:      vlan["dhcp"].(bool),
					VlanID:        uint16(vlan["id"].(int)),
					VlanMTU:       uint32(vlan["mtu"].(int)),
				})
			}
		}

		devices = append(devices, dev)
	}

	return
}

// ExpandProxyConfig safely extracts values from a map of string keys denoting proxy server arguments into a pointer to a Talos ProxyConfig.
// TODO change argument type to handle a whole ProxyConfig.
// More info https://www.talos.dev/v1.0/reference/configuration/#proxyconfig
func ExpandProxyConfig(proxyArgs TypeMap) (proxyConfig *v1alpha1.ProxyConfig, err error) {
	proxyConfig = &v1alpha1.ProxyConfig{}

	proxyConfig.ExtraArgsConfig = map[string]string{}
	for k, v := range proxyArgs {
		proxyConfig.ExtraArgsConfig[k] = v.(string)
	}

	return
}

// ExpandProxyConfig safely extracts values from a map of string keys denoting API server arguments into a pointer to a Talos APIServerConfig.
// TODO change input argument type to handle a whole APIServerConfig, merge with certSANs handler.
// More info https://www.talos.dev/v1.0/reference/configuration/#proxyconfig
func ExpandAPIServerConfig(apiServerArgs TypeMap) (apiServerConfig *v1alpha1.APIServerConfig, err error) {
	apiServerConfig = &v1alpha1.APIServerConfig{}

	apiServerConfig.ExtraArgsConfig = map[string]string{}
	for k, v := range apiServerArgs {
		apiServerConfig.ExtraArgsConfig[k] = v.(string)
	}

	return
}

// generateCommonConfig gets values from the schema's resourcedata and passes them into Talos's config data structure
// for the purpose of node configuration file generation.
func generateCommonConfig(d *schema.ResourceData, config *v1alpha1.Config) diag.Diagnostics {
	mc := config.MachineConfig
	cc := config.ClusterConfig

	// Install configuration
	mc.MachineInstall.InstallDisk = d.Get("install_disk").(string)
	mc.MachineInstall.InstallImage = d.Get("talos_image").(string)
	mc.MachineInstall.InstallExtraKernelArgs = ExpandTypeList[string](d.Get("kernel_args").(TypeList))

	var err error
	if cc.ProxyConfig, err = ExpandProxyConfig(d.Get("cluster_proxy_args").(TypeMap)); err != nil {
		return diag.FromErr(err)
	}

	if cc.APIServerConfig, err = ExpandAPIServerConfig(d.Get("cluster_apiserver_args").(TypeMap)); err != nil {
		return diag.FromErr(err)
	}

	// Network configuration
	mc.MachineNetwork.NetworkHostname = d.Get("name").(string)
	mc.MachineNetwork.NameServers = ExpandTypeList[string](d.Get("nameservers").(TypeList))

	ifs, err := ExpandDeviceList(d.Get("interface").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}
	mc.MachineNetwork.NetworkInterfaces = ifs

	interfaces_ := d.Get("interface").([]interface{})
	for i, iface := range mc.MachineNetwork.NetworkInterfaces {
		// If the device's wireguard configuration exists, derive the public key from it's private key.
		if iface.DeviceWireguardConfig != nil {
			priv := iface.DeviceWireguardConfig.WireguardPrivateKey

			var pk wgtypes.Key
			if pk, err = wgtypes.ParseKey(priv); err != nil {
				return diag.FromErr(err)
			}

			interfaces_[i].(map[string]interface{})["wireguard"].([]interface{})[0].(map[string]interface{})["public_key"] =
				pk.PublicKey().String()
			interfaces_[i].(map[string]interface{})["wireguard"].([]interface{})[0].(map[string]interface{})["private_key"] =
				priv
		}
	}
	// Set the value of the schema's interface field to the previously modified network interface list.
	if err := d.Set("interface", interfaces_); err != nil {
		return diag.FromErr(err)
	}

	for _, mount := range d.Get("kubelet_extra_mount").([]interface{}) {
		m := mount.(map[string]interface{})

		mountOptions := []string{}
		for _, option := range m["options"].([]interface{}) {
			mountOptions = append(mountOptions, option.(string))
		}

		mc.MachineKubelet.KubeletExtraMounts = append(mc.MachineKubelet.KubeletExtraMounts, v1alpha1.ExtraMount{
			Mount: specs.Mount{
				Destination: m["destination"].(string),
				Type:        m["type"].(string),
				Source:      m["source"].(string),
				Options:     mountOptions,
			},
		})
	}
	/*
		mc.MachineRegistries.RegistryMirrors = GetTypeMapWrapEach(d.Get("registry_mirrors"), func(v string) *v1alpha1.RegistryMirrorConfig {
			return &v1alpha1.RegistryMirrorConfig{
				MirrorEndpoints: []string{v},
			}
		})
	*/
	mc.MachineKubelet.KubeletExtraArgs = ExpandTypeMap[string](d.Get("kubelet_extra_args").(TypeMap))
	mc.MachineSysctls = ExpandTypeMap[string](d.Get("sysctls").(TypeMap))

	return nil
}

package talos

import (
	"fmt"
	"net"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func validateDomain(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain, got \"%s\"", key, v))
	}
	return
}

func validateMAC(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, err := net.ParseMAC(v); err != nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid MAC address, got \"%s\", error \"%s\"", key, v, err.Error()))
	}
	return
}

func validateCIDR(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, _, err := net.ParseCIDR(v); err != nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid CIDR IP address, got \"%s\", error \"%s\"", key, v, err.Error()))
	}
	return
}

func validateIP(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if net.ParseIP(v) == nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid IP address, got \"%s\"", key, v))
	}
	return
}

func validateHost(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*:[0-9]{2,}`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain with a port appended, seperated by \":\", got \"%s\"", key, v))
	}
	return
}

func validateEndpoint(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(:[0-9]{2,})?`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain with an optional port appended, seperated by \":\", got \"%s\"", key, v))
	}
	return
}

func validateImage(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[^/]+\.[^/.]+/([^/.]+/)?[^/.]+(:.+)?`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a valid container image, got \"%s\"", key, v))
	}
	return
}

func validateKey(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, err := wgtypes.ParseKey(v); err != nil {
		errs = append(errs, err)
	}
	return
}

var (
	kubeletExtraMountSchema schema.Schema = schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination": {
					Type:     schema.TypeString,
					Required: true,
				},
				"type": {
					Type:     schema.TypeString,
					Optional: true,
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
					Type:     schema.TypeString,
					Required: true,
					//ValidateFunc: validatePath,
				},
				"options": {
					Type:     schema.TypeList,
					Optional: true,
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

	networkInterfaceSchema schema.Schema = schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"addresses": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validateCIDR,
					},
				},
				"wireguard": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"peer": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"allowed_ips": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validateCIDR,
											},
										},
										"endpoint": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validateEndpoint,
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
											Default: 0,
										},
										"public_key": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validateKey,
										},
									},
								},
							},
							"public_key": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"private_key": {
								Type:         schema.TypeString,
								Sensitive:    true,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validateKey,
							},
						},
					},
				},
				"route": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"network": {
								Type:         schema.TypeString,
								Optional:     true,
								InputDefault: "0.0.0.0/0",
								ValidateFunc: validateCIDR,
							},
							"gateway": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validateIP,
							},
						},
					},
				},
			},
		},
	}
)

package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	machinetype "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func validateGpu(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	switch v {
	case
		"Cometlake",
		"AnyGPU":
	default:
		errs = append(errs, fmt.Errorf("Invalid keepalived node state, expected one of Cometlake, AnyGPU, got %s", v))
	}

	return
}

func resourceWorkerNode() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkerNodeCreate,
		ReadContext:   resourceWorkerNodeRead,
		UpdateContext: resourceWorkerNodeUpdate,
		DeleteContext: resourceWorkerNodeDelete,
		Schema: map[string]*schema.Schema{
			// Mandatory for minimal template generation
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"install_disk": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kernel_args": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cluster_apiserver_args": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cluster_proxy_args": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"talos_image": {
				Type:     schema.TypeString,
				Required: true,
			},
			"macaddr": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dhcp_network_cidr": {
				Type:     schema.TypeString,
				Required: true,
			},
			"registry_mirrors": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"kubelet_extra_mount": {
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
							Required: true,
						},
						"source": {
							Type:     schema.TypeString,
							Required: true,
						},
						"options": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"kubelet_extra_args": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sysctls": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"udev": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"interface": {
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
								Type: schema.TypeString,
							},
						},
						"route": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network": {
										Type:         schema.TypeString,
										Optional:     true,
										InputDefault: "0.0.0.0/0",
									},
									"gateway": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"nameservers": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// From the cluster provider
			"base_config": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			// Generated
			"patch": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

type WorkerNodeSpec struct {
	Name string

	IPNetwork   string
	Hostname    string
	Gateway     string
	Nameservers []string

	Privileged bool
	GPU        string
	Mayastor   bool

	RegistryIP string
}

func assignSchemaStringList(d *schema.ResourceData, field string, configField *[]string) {
	for _, value := range d.Get(field).([]interface{}) {
		*configField = append(*configField, value.(string))
	}

}

func generateConfigWorker(ctx context.Context, d *schema.ResourceData) ([]byte, diag.Diagnostics) {
	input := generate.Input{}
	if err := json.Unmarshal([]byte(d.Get("base_config").(string)), &input); err != nil {
		tflog.Error(ctx, "Failed to unmarshal input bundle: "+err.Error())
		return nil, diag.FromErr(err)
	}

	var workerConfig *v1alpha1.Config
	workerConfig, err := generate.Config(machinetype.TypeWorker, &input)
	if err != nil {
		tflog.Error(ctx, "failed to generate config for node: "+err.Error())
		return nil, diag.FromErr(err)
	}

	mc := workerConfig.MachineConfig
	// Install opts
	mc.MachineInstall.InstallDisk = d.Get("install_disk").(string)
	mc.MachineInstall.InstallImage = d.Get("talos_image").(string)
	assignSchemaStringList(d, "kernel_args", &mc.MachineInstall.InstallExtraKernelArgs)

	// Network opts
	mc.MachineNetwork.NetworkHostname = d.Get("name").(string)
	assignSchemaStringList(d, "nameservers", &mc.MachineNetwork.NameServers)

	interfaces := d.Get("interface").([]interface{})
	for _, netInterface := range interfaces {
		n := netInterface.(map[string]interface{})

		addresses := []string{}
		for _, value := range n["addresses"].([]interface{}) {
			addresses = append(addresses, value.(string))
		}

		routes := []*v1alpha1.Route{}
		for _, resourceRoute := range n["route"].([]interface{}) {
			r := resourceRoute.(map[string]interface{})

			routes = append(routes, &v1alpha1.Route{
				RouteGateway: r["gateway"].(string),
				RouteNetwork: r["network"].(string),
			})
		}

		mc.MachineNetwork.NetworkInterfaces = append(mc.MachineNetwork.NetworkInterfaces, &v1alpha1.Device{
			DeviceInterface: n["name"].(string),
			DeviceAddresses: addresses,
			DeviceRoutes:    routes,
		})
	}

	// TODO: Model full capabilities
	/*
		for host, endpoint := range d.Get("registry_mirrors").(map[string]interface{}) {
			mc.MachineRegistries.RegistryMirrors = {
			host:
			}
			mc.MachineRegistries.RegistryMirrors[host].MirrorEndpoints[0] = endpoint.(string)
		}
	*/

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

	mc.MachineKubelet.KubeletExtraArgs = map[string]string{}
	for k, v := range d.Get("kubelet_extra_args").(map[string]interface{}) {
		mc.MachineKubelet.KubeletExtraArgs[k] = v.(string)
	}

	mc.MachineSysctls = map[string]string{}
	for k, v := range d.Get("sysctls").(map[string]interface{}) {
		mc.MachineSysctls[k] = v.(string)
	}

	udevRules := []string{}
	for _, v := range d.Get("udev").([]interface{}) {
		udevRules = append(udevRules, v.(string))
	}

	mc.MachineUdev = &v1alpha1.UdevConfig{
		UdevRules: udevRules,
	}

	apiExtraArgs := map[string]string{}
	for k, v := range d.Get("cluster_apiserver_args").(map[string]interface{}) {
		apiExtraArgs[k] = v.(string)
	}
	workerConfig.ClusterConfig.APIServerConfig = &v1alpha1.APIServerConfig{}

	proxyExtraArgs := map[string]string{}
	for k, v := range d.Get("cluster_proxy_args").(map[string]interface{}) {
		proxyExtraArgs[k] = v.(string)
	}
	workerConfig.ClusterConfig.ProxyConfig = &v1alpha1.ProxyConfig{}

	var workerYaml []byte

	workerYaml, err = workerConfig.Bytes()
	if err != nil {
		log.Fatalf("failed to generate config" + err.Error())
		return nil, diag.FromErr(err)
	}

	return workerYaml, nil
}

func resourceWorkerNodeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	patched, diags := generateConfigWorker(ctx, d)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}

	/*
		patched, diags := generatePatchedWorker(ctx, d, cfg)
		if diags != nil {
			tflog.Error(ctx, "Error generating patched machineconfig")
			return diags
		}
		d.Set("patch", patched)
	*/
	_, network, err := net.ParseCIDR(d.Get("dhcp_network_cidr").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	mac, err := net.ParseMAC(d.Get("macaddr").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	ip, diags := lookupIP(ctx, network, mac)
	if diags != nil {
		tflog.Error(ctx, "Error looking up node IP")
		return diags
	}

	talosport := 50000
	host := ip.String() + ":" + strconv.Itoa(talosport)

	tflog.Error(ctx, "Waiting for talos machine")
	tlsConfig, diags := makeTlsConfig(generate.Certs{}, false)
	if diags != nil {
		return diags
	}
	waitTillTalosMachineUp(ctx, tlsConfig, host, false)
	tflog.Error(ctx, "finished waiting for talos machine")

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)),
	}
	conn, err := grpc.DialContext(ctx, host, opts...)
	defer conn.Close()
	if err != nil {
		tflog.Error(ctx, "Error dialing talos GRPC endpoint.")
		return diag.FromErr(err)
	}

	client := machine.NewMachineServiceClient(conn)
	_, err = client.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: []byte(patched),
		Mode: machine.ApplyConfigurationRequest_Mode(machine.ApplyConfigurationRequest_REBOOT),
	})
	if err != nil {
		tflog.Error(ctx, "Error applying configuration")
		tflog.Error(ctx, err.Error())
		return diag.FromErr(err)
	}

	d.SetId(d.Get("name").(string))
	d.Set("patch", patched)

	return nil
}

func resourceWorkerNodeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
func resourceWorkerNodeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
func resourceWorkerNodeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

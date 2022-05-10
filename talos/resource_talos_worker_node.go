package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"kubelet_extra_mount": &kubeletExtraMountSchema,
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
			"interface": &networkInterfaceSchema,
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

	udevRules := []string{}
	for _, v := range d.Get("udev").([]interface{}) {
		udevRules = append(udevRules, v.(string))
	}

	mc.MachineUdev = &v1alpha1.UdevConfig{
		UdevRules: udevRules,
	}

	generateCommonConfig(d, workerConfig)

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

	network := d.Get("dhcp_network_cidr").(string)
	mac := d.Get("macaddr").(string)

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
		Data: patched,
		Mode: machine.ApplyConfigurationRequest_Mode(machine.ApplyConfigurationRequest_REBOOT),
	})
	if err != nil {
		tflog.Error(ctx, "Error applying configuration")
		tflog.Error(ctx, err.Error())
		return diag.FromErr(err)
	}

	d.SetId(d.Get("name").(string))
	d.Set("patch", string(patched))

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

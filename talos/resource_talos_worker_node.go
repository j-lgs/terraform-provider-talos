package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"text/template"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/config/configpatcher"
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
			"ip": {
				Type:     schema.TypeString,
				Required: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Required: true,
			},
			"nameservers": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// Container registry optionals
			"registry_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"privileged": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"mayastor": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"gpu": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateGpu,
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

func generatePatchedWorker(ctx context.Context, d *schema.ResourceData, config []byte) (string, diag.Diagnostics) {
	nameservers := []string{}
	for _, ns := range d.Get("nameservers").([]interface{}) {
		nameservers = append(nameservers, ns.(string))
	}

	var t *template.Template

	// template controlplane patches
	t = template.Must(template.New("worker").Parse(templateWorker()))

	ip, network, err := net.ParseCIDR(d.Get("ip").(string))
	if err != nil {
		return "", diag.FromErr(err)
	}

	buffer := new(strings.Builder)
	err = t.ExecuteTemplate(buffer, "worker", WorkerNodeSpec{
		Name: d.Get("name").(string),

		IPNetwork:   ipNetwork(ip, *network),
		Hostname:    d.Get("name").(string),
		Gateway:     d.Get("gateway").(string),
		Nameservers: nameservers,

		Privileged: d.Get("privileged").(bool),
		Mayastor:   d.Get("mayastor").(bool),
		GPU:        d.Get("gpu").(string),

		RegistryIP: d.Get("registry_ip").(string),
	})
	if err != nil {
		tflog.Error(ctx, "Error running worker template.")
		return "", diag.FromErr(err)
	}

	jsonpatch, err := jsonpatch.DecodePatch([]byte(buffer.String()))
	if err != nil {
		tflog.Error(ctx, "Error decoding jsonpatch: "+buffer.String())
		return "", diag.FromErr(err)
	}

	patched, err := configpatcher.JSON6902(config, jsonpatch)
	if err != nil {
		tflog.Error(ctx, "Error attempting applying jsonpatch: "+buffer.String())
		return "", diag.FromErr(err)
	}

	return string(patched), nil
}

func generateConfigWorker(ctx context.Context, d *schema.ResourceData) ([]byte, diag.Diagnostics) {
	disk := d.Get("install_disk").(string)
	image := d.Get("talos_image").(string)

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

	workerConfig.MachineConfig.MachineInstall.InstallDisk = disk
	workerConfig.MachineConfig.MachineInstall.InstallImage = image
	var workerYaml []byte

	workerYaml, err = workerConfig.Bytes()
	if err != nil {
		log.Fatalf("failed to generate config" + err.Error())
		return nil, diag.FromErr(err)
	}

	return workerYaml, nil
}

func resourceWorkerNodeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cfg, diags := generateConfigWorker(ctx, d)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}

	patched, diags := generatePatchedWorker(ctx, d, cfg)
	if diags != nil {
		tflog.Error(ctx, "Error generating patched machineconfig")
		return diags
	}
	d.Set("patch", patched)

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

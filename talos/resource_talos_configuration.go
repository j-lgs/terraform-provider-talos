package talos

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"hash/fnv"

	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Schema: map[string]*schema.Schema{
			"target_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The version of the Talos cluster configuration that will be generated.",
			},
			"cluster_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Configures the cluster's name",
			},
			"endpoints": {
				Type:     schema.TypeList,
				MinItems: 1,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of that the talosctl client will connect to. Can be a DNS hostname or an IP address and may include a port number. Must begin with \"https://\".",
			},
			"kubernetes_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The kubernetes endpoint that the nodes and the kubectl client will connect to. Can be a DNS hostname or an IP address and may include a port number. Must begin with \"https://\".",
			},
			"kubernetes_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The version of kubernetes and all it's components (kube-apiserver, kubelet, kube-scheduler, etc) that will be deployed onto the cluster.",
			},

			"talosconfig": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
				Description: "Talosconfig YAML that can be used by the talosctl client to communicate with the cluster.",
			},
			"base_config": {
				Sensitive:   true,
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON Serialised object that contains information needed to create controlplane and worker node configurations.",
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	targetVersion := d.Get("target_version").(string)
	kubernetesVersion := d.Get("kubernetes_version").(string)
	clusterName := d.Get("cluster_name").(string)
	endpoint := d.Get("kubernetes_endpoint").(string)

	var (
		versionContract = config.TalosVersionCurrent //nolint:wastedassign,ineffassign
		err             error
	)

	versionContract, err = config.ParseContractFromVersion(targetVersion)
	if err != nil {
		tflog.Error(ctx, "failed to parse version contract: "+err.Error())
		return diag.FromErr(err)
	}

	secrets, err := generate.NewSecretsBundle(generate.NewClock(), generate.WithVersionContract(versionContract))
	if err != nil {
		tflog.Error(ctx, "failed to generate secrets bundle: "+err.Error())
		return diag.FromErr(err)
	}

	endpointList := d.Get("endpoints").([]interface{})
	endpoints := []string{}
	for _, endpoint := range endpointList {
		endpoints = append(endpoints, endpoint.(string))
	}

	input, err := generate.NewInput(clusterName, endpoint, kubernetesVersion, secrets,
		generate.WithVersionContract(versionContract),
	)
	if err != nil {
		tflog.Error(ctx, "Error generating input.")
		return diag.FromErr(err)
	}
	input_json, err := json.Marshal(input)
	if err != nil {
		tflog.Error(ctx, "failed to unmarshal secrets bundle: "+err.Error())
		return diag.FromErr(err)
	}
	d.Set("base_config", string(input_json))

	talosconfig, err := generate.Talosconfig(input, generate.WithEndpointList(endpoints))
	if err != nil {
		tflog.Error(ctx, "Error generating talosconfig.")
		return diag.FromErr(err)
	}

	config, err := yaml.Marshal(talosconfig)
	if err != nil {
		tflog.Error(ctx, "Error getting talosconfig bytes.")
		return diag.FromErr(err)
	}
	d.Set("talosconfig", string(config))

	hash := fnv.New128().Sum([]byte(clusterName))
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(b64, hash)

	d.SetId(string(b64))

	return nil
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

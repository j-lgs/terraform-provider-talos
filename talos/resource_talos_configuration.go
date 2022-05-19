package talos

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"hash/fnv"

	"github.com/ghodss/yaml"
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"

	//	"github.com/hashicorp/terraform-plugin-log/tflog"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfsdk.ResourceType = talosClusterConfigResourceType{}
var _ tfsdk.Resource = talosClusterConfigResource{}
var _ tfsdk.ResourceWithImportState = talosClusterConfigResource{}

type talosClusterConfigResourceType struct{}

func (t talosClusterConfigResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Represents the basic CA/CRT bundle that's needed to provision a Talos cluster. Contains information that is shared with, and essential for the creation of, worker and controlplane nodes.",

		Attributes: map[string]tfsdk.Attribute{
			"target_version": {
				MarkdownDescription: "The version of the Talos cluster configuration that will be generated.",
				Required:            true,
				Type:                types.StringType,
			},
			"name": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "Configures the cluster's name",
			},
			"talos_endpoints": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required:            true,
				MarkdownDescription: "A list of that the talosctl client will connect to. Can be a DNS hostname or an IP address and may include a port number. Must begin with \"https://\".",
			},
			"kubernetes_endpoint": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "The kubernetes endpoint that the nodes and the kubectl client will connect to. Can be a DNS hostname or an IP address and may include a port number. Must begin with \"https://\".",
			},
			"kubernetes_version": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "The version of kubernetes and all it's components (kube-apiserver, kubelet, kube-scheduler, etc) that will be deployed onto the cluster.",
			},

			"talos_config": {
				Type:                types.StringType,
				Sensitive:           true,
				Computed:            true,
				MarkdownDescription: "Talosconfig YAML that can be used by the talosctl client to communicate with the cluster.",
			},
			"base_config": {
				Sensitive:           true,
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "JSON Serialised object that contains information needed to create controlplane and worker node configurations.",
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Identifier hash, derived from the cluster's name.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t talosClusterConfigResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return talosClusterConfigResource{
		provider: provider,
	}, diags
}

type talosClusterConfigResourceData struct {
	TargetVersion      types.String `tfsdk:"target_version"`
	ClusterName        types.String `tfsdk:"name"`
	Endpoints          types.List   `tfsdk:"talos_endpoints"`
	KubernetesEndpoint types.String `tfsdk:"kubernetes_endpoint"`
	KubernetesVersion  types.String `tfsdk:"kubernetes_version"`
	TalosConfig        types.String `tfsdk:"talos_config"`
	BaseConfig         types.String `tfsdk:"base_config"`
	Id                 types.String `tfsdk:"id"`
}

type talosClusterConfigResource struct {
	provider provider
}

func (r talosClusterConfigResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var (
		versionContract = config.TalosVersionCurrent //nolint:wastedassign,ineffassign
		err             error
		data            talosClusterConfigResourceData
	)

	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos configuration's Create method has been called without the provider being configured. This is a provider bug.")
		return
	}

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetVersion := data.TargetVersion.Value
	kubernetesVersion := data.KubernetesVersion.Value
	clusterName := data.ClusterName.Value
	endpoints := []string{}
	diags = data.Endpoints.ElementsAs(ctx, &endpoints, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err != nil {
		diags.AddError("Unable to generate secrets bundle", err.Error())
		resp.Diagnostics.Append(diags...)
	}

	versionContract, err = config.ParseContractFromVersion(targetVersion)
	if err != nil {
		diags.AddError("Unable to parse versionContract", err.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	secrets, err := generate.NewSecretsBundle(generate.NewClock(), generate.WithVersionContract(versionContract))
	if err != nil {
		diags.AddError("Unable to generate secrets bundle", err.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	input, err := generate.NewInput(clusterName, data.KubernetesEndpoint.Value, kubernetesVersion, secrets,
		generate.WithVersionContract(versionContract),
	)
	if err != nil {
		diags.AddError("Error generating input bundle", err.Error())
		return
	}
	input_json, err := json.Marshal(input)
	if err != nil {
		diags.AddError("failed to unmarshal to secrets bundle to a json string: ", err.Error())
		return
	}
	data.BaseConfig = types.String{Value: string(input_json)}

	talosconfig, err := generate.Talosconfig(input, generate.WithEndpointList(endpoints))
	if err != nil {
		diags.AddError("Error generating talosconfig.", err.Error())
		return
	}

	config, err := yaml.Marshal(talosconfig)
	if err != nil {
		diags.AddError("Error getting talosconfig bytes.", err.Error())
		return
	}
	data.TalosConfig = types.String{Value: string(config)}

	hash := fnv.New128().Sum([]byte(clusterName))
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(b64, hash)

	data.Id = types.String{Value: string(b64)}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	return
}

func (r talosClusterConfigResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos configuration's Read method has been called without the provider being configured. This is a provider bug.")
		return
	}
}

func (r talosClusterConfigResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos configuration's Update method has been called without the provider being configured. This is a provider bug.")
		return
	}
}

func (r talosClusterConfigResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos configuration's Delete method has been called without the provider being configured. This is a provider bug.")
		return
	}
}

func (r talosClusterConfigResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

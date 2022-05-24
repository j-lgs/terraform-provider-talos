package talos

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"hash/fnv"

	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
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
				Type:     types.StringType,
				Optional: true,
				MarkdownDescription: `The canonical address of the kubernetes control plane.
						It can be a DNS name, the IP address of a load balancer, or (default) the IP address of the
						first master node.  It is NOT multi-valued.  It may optionally specify the port.`,
			},
			"secret_bundle": {
				Optional:    true,
				Description: SecretBundleSchema.MarkdownDescription,
				Attributes:  tfsdk.SingleNestedAttributes(SecretBundleSchema.Attributes),
			},
			"k8s_cert_sans": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"machine_cert_sans": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"service_domain": {
				Type:     types.StringType,
				Optional: true,
			},
			"pod_network": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"service_network": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"kubernetes_version": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "The version of kubernetes and all it's components (kube-apiserver, kubelet, kube-scheduler, etc) that will be deployed onto the cluster.",
			},
			"external_etcd": {
				Type:     types.BoolType,
				Optional: true,
			},
			"install_disk": {
				Type:     types.StringType,
				Optional: true,
			},
			"install_image": {
				Type:     types.StringType,
				Optional: true,
				// TODO validate
				// ValidateFunc: validateImage,
			},
			"install_extra_kargs": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"network": {
				Optional:            true,
				MarkdownDescription: NetworkConfigOptionSchema.MarkdownDescription,
				Attributes:          tfsdk.ListNestedAttributes(NetworkConfigOptionSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			},
			"cni": {
				Optional:    true,
				Description: "Choose cluster CNI. one of flannel, custom and none",
				Type:        types.StringType,
			},
			"registry": {
				Optional:    true,
				Description: RegistrySchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(RegistrySchema.Attributes),
			},
			"disks": {
				Optional:    true,
				Description: MachineDiskSchema.MarkdownDescription,
				Attributes:  tfsdk.ListNestedAttributes(MachineDiskSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			},
			"encryption": {
				Optional:    true,
				Description: EncryptionSchema.MarkdownDescription,
				Attributes:  tfsdk.SingleNestedAttributes(EncryptionSchema.Attributes),
			},
			"sysctls": {
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Description: "Used to configure the machine’s sysctls.",
			},
			"debug": {
				Type:     types.BoolType,
				Optional: true,
			},
			"persist": {
				Type:     types.BoolType,
				Optional: true,
			},
			"allow_scheduling_on_masters": {
				Type:     types.BoolType,
				Optional: true,
			},
			"discovery": {
				Type:     types.BoolType,
				Optional: true,
			},

			// Generated
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
	TargetVersion      types.String    `tfsdk:"target_version"`
	ClusterName        types.String    `tfsdk:"name"`
	Endpoints          []types.String  `tfsdk:"talos_endpoints"`
	KubernetesEndpoint types.String    `tfsdk:"kubernetes_endpoint"`
	KubernetesVersion  types.String    `tfsdk:"kubernetes_version"`
	Encryption         *EncryptionData `tfsdk:"encryption"`

	TalosConfig types.String `tfsdk:"talos_config"`
	BaseConfig  types.String `tfsdk:"base_config"`
	ID          types.String `tfsdk:"id"`
	Disks                    []MachineDiskData       `tfsdk:"disks"`
}

func (plan *talosClusterConfigResourceData) Generate(opts []generate.GenOption) (err error) {
	var versionContract = config.TalosVersionCurrent //nolint:wastedassign,ineffassign

	targetVersion := plan.TargetVersion.Value
	kubernetesVersion := plan.KubernetesVersion.Value
	clusterName := plan.ClusterName.Value

	versionContract, err = config.ParseContractFromVersion(targetVersion)
	if err != nil {
		return fmt.Errorf("unable to parse version contract: %w", err)
	}

	secrets, err := generate.NewSecretsBundle(generate.NewClock(), generate.WithVersionContract(versionContract))
	if err != nil {
		return fmt.Errorf("unable to generate secrets bundle: %w", err)
	}

	input, err := generate.NewInput(clusterName, plan.KubernetesEndpoint.Value, kubernetesVersion, secrets,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("error generating input bundle: %w", err)
	}

	//lint:ignore SA1026 suppress check as it's issue is with a datastructure outside the project's scope
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to secrets bundle to a JSON string: %w", err)
	}
	plan.BaseConfig = types.String{Value: string(inputJSON)}

	endpoints := []string{}
	for _, endpoint := range plan.Endpoints {
		endpoints = append(endpoints, endpoint.Value)
	}
	talosconfig, err := generate.Talosconfig(input, generate.WithEndpointList(endpoints))
	if err != nil {
		return fmt.Errorf("error generating talosconfig: %w", err)
	}

	config, err := talosconfig.Bytes()
	if err != nil {
		return fmt.Errorf("error getting talosconfig bytes: %w", err)
	}

	plan.TalosConfig = types.String{Value: string(config)}
	return
}

func (plan *talosClusterConfigResourceData) TalosData() (out []generate.GenOption, err error) {
	var versionContract = config.TalosVersionCurrent //nolint:wastedassign,ineffassign

	out = []generate.GenOption{}

	if plan.Encryption != nil {
		encryption, err := plan.Encryption.Data()
		if err != nil {
			return nil, err
		}
		out = append(out, generate.WithSystemDiskEncryption(encryption.(*v1alpha1.SystemDiskEncryptionConfig)))
	}

	generate.WithVersionContract(versionContract)

	return
}

func (plan *talosClusterConfigResourceData) ReadInto(in *generate.Input) (err error) {
	if in.SystemDiskEncryptionConfig != nil {
		if plan.Encryption != nil {
			err := plan.Encryption.Read(in.SystemDiskEncryptionConfig)
			if err != nil {
				return err
			}
		}
	}

	return
}

type talosClusterConfigResource struct {
	provider provider
}

func (r talosClusterConfigResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var (
		err  error
		data talosClusterConfigResourceData
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

	genopts, err := data.TalosData()
	if err != nil {
		resp.Diagnostics.AddError("unable to get TalosData from plan", err.Error())
		return
	}

	data.Generate(genopts)

	clusterName := data.ClusterName.Value
	hash := fnv.New128().Sum([]byte(clusterName))
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(b64, hash)

	data.ID = types.String{Value: string(b64)}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r talosClusterConfigResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos configuration's Read method has been called without the provider being configured. This is a provider bug.")
	}
}

func (r talosClusterConfigResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos configuration's Update method has been called without the provider being configured. This is a provider bug.")
	}
}

func (r talosClusterConfigResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos configuration's Delete method has been called without the provider being configured. This is a provider bug.")
	}
}

func (r talosClusterConfigResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

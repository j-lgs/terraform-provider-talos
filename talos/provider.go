package talos

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	talosPort = 50000
)

var _ tfsdk.Provider = &provider{}

type provider struct {
	configured bool
	// skipdelete and skipread are used to recover from broken deployments, where the node fails to deploy,
	// but the provider thinks it exists. In this situation it will try to refresh its status using talos's api,
	// which will not be up because nodes are in a broken state. This will cause the plugin to hang and timeout when
	// connecting.
	skipdelete bool
	skipread   bool
	version    string
}

// Configure creates an instance of a Talos API helper struct and set it as the "client" attribute for the provider struct.
func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	if resp.Diagnostics.HasError() {
		return
	}

	if val, set := os.LookupEnv("TALOS_SKIPDELETE"); set {
		b, err := strconv.ParseBool(val)
		if err != nil {
			resp.Diagnostics.AddError("environment parse error",
				"error parsing boolean value for TALOS_SKIPDELETE")
		}

		p.skipdelete = b
	}

	if val, set := os.LookupEnv("TALOS_SKIPREAD"); set {
		b, err := strconv.ParseBool(val)
		if err != nil {
			resp.Diagnostics.AddError("environment parse error",
				"error parsing boolean value for TALOS_SKIPREAD")
		}

		p.skipread = b
	}

	p.configured = true
}

// GetResources returns a map of all provider resources.
func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"talos_configuration": talosClusterConfigResourceType{},
		"talos_control_node":  talosControlNodeResourceType{},
		"talos_worker_node":   talosWorkerNodeResourceType{},
	}, nil
}

// GetDataSources is a stub for implementing the terraform provider interface
func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"example": {
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

// New returns a function that creates a new instance of the provider struct
func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
// Taken verbatim from the terraform provider scaffold. available here:
// https://github.com/hashicorp/terraform-provider-scaffolding-framework
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}

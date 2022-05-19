package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/talos-systems/talos/pkg/machinery/api/resource"

	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"

	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"gopkg.in/yaml.v2"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	machinetype "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ tfsdk.ResourceType = talosWorkerNodeResourceType{}
var _ tfsdk.Resource = talosWorkerNodeResource{}
var _ tfsdk.ResourceWithImportState = talosWorkerNodeResource{}

type talosWorkerNodeResourceType struct{}

func (t talosWorkerNodeResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Represents a Talos worker node.",
		Attributes: map[string]tfsdk.Attribute{
			// Mandatory for minimal template generation
			"name": {
				Type:     types.StringType,
				Required: true,
				// ValidateFunc: validateDomain,
				// ForceNew: true,
				// TODO validate and fix forcenew
			},
			// Install arguments
			"install_disk": {
				Type:     types.StringType,
				Required: true,
			},
			"talos_image": {
				Type:     types.StringType,
				Required: true,
				// TODO validate
				// ValidateFunc: validateImage,
			},
			"kernel_args": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"macaddr": {
				Type:     types.StringType,
				Required: true,
				// TODO validate and forcenew
				// ForceNew: true,
				// ValidateFunc: validateMAC,
			},
			"dhcp_network_cidr": {
				Type:     types.StringType,
				Required: true,
				// TODO validate
				// ValidateFunc: validateCIDR,
			},
			// --- MachineConfig.
			// See https://www.talos.dev/v1.0/reference/configuration/#machineconfig for full spec.

			"cert_sans": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
				// TODO validation
				Description: "Extra certificate subject alternative names for the machine’s certificate.",
			},

			"control_plane": {
				Optional:    true,
				Description: ControlPlaneConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(ControlPlaneConfigSchema.Attributes),
			},

			"kubelet": {
				Optional:    true,
				Description: KubeletConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(KubeletConfigSchema.Attributes),
			},

			"pod": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
				// TODO validation
				Description: "Used to provide static pod definitions to be run by the kubelet directly bypassing the kube-apiserver.",
			},
			// hostname derived from name
			"devices": {
				Required:    true,
				Description: NetworkDeviceSchema.Description,
				Attributes:  tfsdk.MapNestedAttributes(NetworkDeviceSchema.Attributes, tfsdk.MapNestedAttributesOptions{}),
			},
			"nameservers": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
				// TODO validation
				// validateEndpoint
				Description: "Used to statically set the nameservers for the machine.",
			},
			"extra_host": {
				Type: types.MapType{
					ElemType: types.ListType{
						ElemType: types.StringType,
					},
				},
				Optional:    true,
				Description: "Allows the addition of user specified files.",
				// TODO validate
			},
			// kubespan not implemented
			// disks not implemented
			// install not implemented
			"files": {
				Optional:    true,
				Description: FileSchema.Description,
				Attributes:  tfsdk.ListNestedAttributes(FileSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			},

			"env": {
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Description: "Allows for the addition of environment variables. All environment variables are set on PID 1 in addition to every service.",
			},
			// time not implemented
			"sysctls": {
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Description: "Used to configure the machine’s sysctls.",
			},
			"sysfs": {
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Description: "Used to configure the machine’s sysctls.",
			},

			"registry": {
				Optional:    true,
				Description: RegistrySchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(RegistrySchema.Attributes),
			},

			// system_disk_encryption not implemented
			// features not implemented
			"udev": {
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Description: "Configures the udev system.",
				Optional:    true,
			},

			// logging not implemented
			// kernel not implemented
			// ----- MachineConfig End
			// ----- Resource Cluster bootstrap configuration
			"bootstrap": {
				Type:     types.BoolType,
				Required: true,
			},
			"bootstrap_ip": {
				Type:     types.StringType,
				Required: true,
				// ValidateFunc: validateIP,
			},

			// From the cluster provider
			"base_config": {
				Type:      types.StringType,
				Required:  true,
				Sensitive: true,
				/*
					ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
						v := value.(string)
						input := generate.Input{}
						if err := json.Unmarshal([]byte(v), &input); err != nil {
							errs = append(errs, fmt.Errorf("Failed to  base_config. Do not set this value to anything other than the base_config value of a talos_cluster_config resource"))
						}
						return
					},
				*/
			},

			// Generated
			"patch": {
				Type:      types.StringType,
				Computed:  true,
				Sensitive: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Identifier hash, derived from the node's name.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type talosWorkerNodeResourceData struct {
	Name            types.String              `tfsdk:"name"`
	InstallDisk     types.String              `tfsdk:"install_disk"`
	TalosImage      types.String              `tfsdk:"talos_image"`
	KernelArgs      map[string]types.String   `tfsdk:"kernel_args"`
	Macaddr         types.String              `tfsdk:"macaddr"`
	DHCPNetworkCidr types.String              `tfsdk:"dhcp_network_cidr"`
	CertSANS        []types.String            `tfsdk:"cert_sans"`
	ControlPlane    *ControlPlaneConfig       `tfsdk:"control_plane"`
	Kubelet         *KubeletConfig            `tfsdk:"kubelet"`
	Pod             []types.String            `tfsdk:"pod"`
	NetworkDevices  map[string]NetworkDevice  `tfsdk:"devices"`
	Nameservers     []types.String            `tfsdk:"nameservers"`
	ExtraHost       map[string][]types.String `tfsdk:"extra_host"`
	Files           []File                    `tfsdk:"files"`
	Env             map[string]types.String   `tfsdk:"env"`
	Proxy           *ProxyConfig              `tfsdk:"proxy"`
	Sysctls         map[string]types.String   `tfsdk:"sysctls"`
	Sysfs           map[string]types.String   `tfsdk:"sysfs"`
	Registry        *Registry                 `tfsdk:"registry"`
	Udev            []types.String            `tfsdk:"udev"`
	ConfigIP        types.String              `tfsdk:"config_ip"`
	BaseConfig      types.String              `tfsdk:"base_config"`
	Patch           types.String              `tfsdk:"patch"`
	Id              types.String              `tfsdk:"id"`
}

func (plan *talosWorkerNodeResourceData) Generate() (err error) {
	return
}

func (plan *talosWorkerNodeResourceData) ReadInto(in *v1alpha1.Config) (err error) {
	return
}

func (plan *talosWorkerNodeResourceData) TalosData(in *v1alpha1.Config) (out *v1alpha1.Config, err error) {
	out = &v1alpha1.Config{}
	in.DeepCopyInto(out)

	md := out.MachineConfig
	cd := out.ClusterConfig
	for _, san := range plan.CertSANS {
		md.MachineCertSANs = append(md.MachineCertSANs, san.Value)
	}

	// Kubelet
	if plan.Kubelet != nil {
		kubelet, err := plan.Kubelet.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		md.MachineKubelet = kubelet.(*v1alpha1.KubeletConfig)
	}

	// NetworkDevices
	md.MachineNetwork = &v1alpha1.NetworkConfig{}
	md.MachineNetwork.NetworkHostname = plan.Name.Value
	md.MachineNetwork.NetworkInterfaces = []*v1alpha1.Device{}
	// set device interfaces after get as it's the map key
	for netInterface, device := range plan.NetworkDevices {
		dev, err := device.Data()
		if err != nil {
			return v1alpha1.Config{}, err
		}
		dev.(*v1alpha1.Device).DeviceInterface = netInterface
		md.MachineNetwork.NetworkInterfaces = append(md.MachineNetwork.NetworkInterfaces, dev.(*v1alpha1.Device))
	}

	md.MachineNetwork.ExtraHostEntries = []*v1alpha1.ExtraHost{}
	for hostname, addresses := range plan.ExtraHost {
		host := &v1alpha1.ExtraHost{
			HostIP: hostname,
		}
		md.MachineNetwork.ExtraHostEntries = append(md.MachineNetwork.ExtraHostEntries, host)
		for _, address := range addresses {
			host.HostAliases = append(host.HostAliases, address.Value)
		}
	}

	md.MachineInstall = &v1alpha1.InstallConfig{
		InstallDisk:  plan.InstallDisk.Value,
		InstallImage: plan.TalosImage.Value,
	}
	if plan.KernelArgs != nil {
		md.MachineInstall.InstallExtraKernelArgs = []string{}
		for k, arg := range plan.KernelArgs {
			md.MachineInstall.InstallExtraKernelArgs = append(md.MachineInstall.InstallExtraKernelArgs, k+"="+arg.Value)
		}
	}

	for _, pod := range plan.Pod {
		var talosPod v1alpha1.Unstructured

		if err = yaml.Unmarshal([]byte(pod.Value), &talosPod); err != nil {
			return
		}

		md.MachinePods = append(md.MachinePods, talosPod)
	}

	if plan.Proxy != nil {
		proxy, err := plan.Proxy.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		cd.ProxyConfig = proxy.(*v1alpha1.ProxyConfig)
	}

	md.MachineEnv = map[string]string{}
	for name, value := range plan.Env {
		md.MachineEnv[name] = value.Value
	}

	for _, planFile := range plan.Files {
		file, err := planFile.Data()
		if err != nil {
			return v1alpha1.Config{}, err
		}
		md.MachineFiles = append(md.MachineFiles, file.(*v1alpha1.MachineFile))
	}

	md.MachineSysctls = map[string]string{}
	for name, value := range plan.Sysctls {
		md.MachineSysctls[name] = value.Value
	}

	md.MachineSysfs = map[string]string{}
	for path, value := range plan.Sysfs {
		md.MachineSysfs[path] = value.Value
	}

	if plan.Registry != nil {
		registries, err := plan.Registry.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		md.MachineRegistries = *registries.(*v1alpha1.RegistriesConfig)
	}

	md.MachineUdev = &v1alpha1.UdevConfig{}
	for _, rule := range plan.Udev {
		md.MachineUdev.UdevRules = append(md.MachineUdev.UdevRules, rule.Value)
	}

	return
}

func (t talosWorkerNodeResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return talosWorkerNodeResource{
		provider: provider,
	}, diags
}

type talosWorkerNodeResource struct {
	provider provider
}

func (r talosWorkerNodeResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var (
		plan talosWorkerNodeResourceData
	)

	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos worker node resource's Create method has been called without the provider being configured. This is a provider bug.")
		return
	}

	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := &plan
	config, errDesc, err := applyConfig(&p, configData{
		Bootstrap:   false,
		CreateNode:  true,
		Mode:        machine.ApplyConfigurationRequest_REBOOT,
		BaseConfig:  plan.BaseConfig.Value,
		MachineType: machinetype.TypeWorker,
		Network:     plan.DHCPNetworkCidr.Value,
		MAC:         plan.Macaddr.Value,
	}, ctx)
	if err != nil {
		resp.Diagnostics.AddError(errDesc, err.Error())
		return
	}

	plan.Patch = types.String{Value: config}

	plan.Id = types.String{Value: string(plan.Name.Value)}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	return
}

func (r talosWorkerNodeResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var (
		state talosWorkerNodeResourceData
	)

	// Error is here. Look into it.
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Error getting plan state.", "")
		return
	}

	if !r.provider.forcedelete {
		conf, errDesc, err := readConfig(&state, readData{
			ConfigIP:   state.ConfigIP.Value,
			BaseConfig: state.BaseConfig.Value,
		}, ctx)
		if err != nil {
			resp.Diagnostics.AddError(errDesc, err.Error())
			return
		}

		state.ReadInto(conf)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	return
}

func (r talosWorkerNodeResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var (
		state talosWorkerNodeResourceData
	)

	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos worker node resource's Update method has been called without the provider being configured. This is a provider bug.")
		return
	}

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := &state
	config, errDesc, err := applyConfig(&p, configData{
		Bootstrap:   false,
		ConfigIP:    state.ConfigIP.Value,
		Mode:        machine.ApplyConfigurationRequest_AUTO,
		BaseConfig:  state.BaseConfig.Value,
		MachineType: machinetype.TypeWorker,
		Network:     state.DHCPNetworkCidr.Value,
		MAC:         state.Macaddr.Value,
	}, ctx)
	if err != nil {
		resp.Diagnostics.AddError(errDesc, err.Error())
		return
	}

	state.Patch = types.String{Value: config}

	if !r.provider.forcedelete {
		conf, errDesc, err := readConfig(&state, readData{
			ConfigIP:   state.ConfigIP.Value,
			BaseConfig: state.BaseConfig.Value,
		}, ctx)
		if err != nil {
			resp.Diagnostics.AddError(errDesc, err.Error())
			return
		}
		state.ReadInto(conf)
	}

	return
}

func (r talosWorkerNodeResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos worker node resource's Read method has been called without the provider being configured. This is a provider bug.")
		return
	}

	return
}

func (r talosWorkerNodeResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

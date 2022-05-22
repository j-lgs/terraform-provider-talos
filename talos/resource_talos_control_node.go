package talos

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"time"

	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"

	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/yaml.v2"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	machinetype "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfsdk.ResourceType = talosControlNodeResourceType{}
var _ tfsdk.Resource = talosControlNodeResource{}
var _ tfsdk.ResourceWithImportState = talosControlNodeResource{}

type talosControlNodeResourceType struct{}

// Note: It will fail on runtime with a Terraform crash if either of required or optional aren't included.
func (t talosControlNodeResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Represents the basic CA/CRT bundle that's needed to provision a Talos cluster. Contains information that is shared with, and essential for the creation of, worker and controlplane nodes.",

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
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Description: "Configures the udev system.",
				Optional:    true,
			},

			// logging not implemented
			// kernel not implemented
			// ----- MachineConfig End
			// ----- ClusterConfig Start

			"control_plane_config": {
				Optional:    true,
				Description: MachineControlPlaneSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(MachineControlPlaneSchema.Attributes),
			},

			// clustername already filled
			// cluster_network not implemented
			"apiserver": {
				Optional:    true,
				Description: APIServerConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(APIServerConfigSchema.Attributes),
			},

			// controller manager not implemented
			"proxy": {
				Optional:    true,
				Description: ProxyConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(ProxyConfigSchema.Attributes),
			},

			// scheduler not implemented
			// discovery not implemented
			// etcd not implemented
			// coredns not implemented
			// external_cloud_provider not implemented
			"extra_manifests": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Description: "A list of urls that point to additional manifests. These will get automatically deployed as part of the bootstrap.",
				Optional:    true,
			},

			// TODO Add verification function confirming it's a correct manifest that can be downloaded.
			// inline_manifests not implemented
			"inline_manifests": {
				Optional:    true,
				Description: InlineManifestSchema.Description,
				Attributes:  tfsdk.ListNestedAttributes(InlineManifestSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			},

			// admin_kubeconfig not implemented
			"allow_scheduling_on_masters": {
				Type:        types.BoolType,
				Optional:    true,
				Description: "Allows running workload on master nodes.",
			},
			// ----- ClusterConfig End

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

type talosControlNodeResourceData struct {
	Name                     types.String              `tfsdk:"name"`
	InstallDisk              types.String              `tfsdk:"install_disk"`
	TalosImage               types.String              `tfsdk:"talos_image"`
	KernelArgs               []types.String            `tfsdk:"kernel_args"`
	Macaddr                  types.String              `tfsdk:"macaddr"`
	DHCPNetworkCidr          types.String              `tfsdk:"dhcp_network_cidr"`
	CertSANS                 []types.String            `tfsdk:"cert_sans"`
	ControlPlane             *ControlPlaneConfig       `tfsdk:"control_plane"`
	Kubelet                  *KubeletConfig            `tfsdk:"kubelet"`
	Pod                      []types.String            `tfsdk:"pod"`
	NetworkDevices           map[string]NetworkDevice  `tfsdk:"devices"`
	Nameservers              []types.String            `tfsdk:"nameservers"`
	ExtraHost                map[string][]types.String `tfsdk:"extra_host"`
	Files                    []File                    `tfsdk:"files"`
	Env                      map[string]types.String   `tfsdk:"env"`
	Sysctls                  map[string]types.String   `tfsdk:"sysctls"`
	Sysfs                    map[string]types.String   `tfsdk:"sysfs"`
	Registry                 *Registry                 `tfsdk:"registry"`
	Udev                     []types.String            `tfsdk:"udev"`
	MachineControlPlane      *MachineControlPlane      `tfsdk:"control_plane_config"`
	APIServer                *APIServerConfig          `tfsdk:"apiserver"`
	Proxy                    *ProxyConfig              `tfsdk:"proxy"`
	ExtraManifests           []types.String            `tfsdk:"extra_manifests"`
	InlineManifests          []InlineManifest          `tfsdk:"inline_manifests"`
	AllowSchedulingOnMasters types.Bool                `tfsdk:"allow_scheduling_on_masters"`
	Bootstrap                types.Bool                `tfsdk:"bootstrap"`
	ConfigIP                 types.String              `tfsdk:"bootstrap_ip"`
	BaseConfig               types.String              `tfsdk:"base_config"`
	Patch                    types.String              `tfsdk:"patch"`
	ID                       types.String              `tfsdk:"id"`
}

func (plan *talosControlNodeResourceData) Generate() (err error) {
	// Generate wireguard keys.
	for _, device := range plan.NetworkDevices {
		// If the device's wireguard configuration exists, derive the public key from it's private key.
		if device.Wireguard != nil {
			var pk wgtypes.Key
			// If a key doesn't exist make one, otherwise generate one.
			if device.Wireguard.PrivateKey.Null {
				pk, err = wgtypes.GeneratePrivateKey()
				device.Wireguard.PrivateKey = types.String{Value: pk.String()}
			} else {
				pk, err = wgtypes.ParseKey(device.Wireguard.PrivateKey.Value)
			}

			if err != nil {
				return err
			}

			device.Wireguard.PublicKey = types.String{Value: pk.PublicKey().String()}
		}
	}

	return
}

func (plan *talosControlNodeResourceData) ReadInto(in *v1alpha1.Config) (err error) {
	plan.Nameservers = []types.String{}
	for _, ns := range in.MachineConfig.MachineNetwork.NameServers {
		plan.Nameservers = append(plan.Nameservers, types.String{Value: ns})
	}
	/*
		plan.ExtraManifests = []types.String{}
		for _, manifestURL := range in.ClusterConfig.ExtraManifests {
			plan.ExtraManifests = append(plan.ExtraManifests, types.String{Value: manifestURL})
		}
		for _, inlineManifest := range in.ClusterConfig.ClusterInlineManifests {
			tfInlineManifest := InlineManifest{}
			err := tfInlineManifest.Read(inlineManifest)
			if err != nil {
				return err
			}
			plan.InlineManifests = append(plan.InlineManifests, tfInlineManifest)
		}
	*/
	return
}

func (plan *talosControlNodeResourceData) TalosData(in *v1alpha1.Config) (out *v1alpha1.Config, err error) {
	out = &v1alpha1.Config{}
	in.DeepCopyInto(out)
	cd := out.ClusterConfig
	if plan.ControlPlane != nil {
		/* Refrain from setting this for now because it's set in the cluster config resource */
		if !plan.ControlPlane.Endpoint.Null {
			url, err := url.Parse(plan.ControlPlane.Endpoint.Value)
			if err != nil {
				return &v1alpha1.Config{}, err
			}
			cd.ControlPlane.Endpoint = &v1alpha1.Endpoint{
				URL: url,
			}
		}
		if !plan.ControlPlane.LocalAPIServerPort.Null {
			cd.ControlPlane.LocalAPIServerPort = int(plan.ControlPlane.LocalAPIServerPort.Value)
		}
	}

	if plan.APIServer != nil {
		apiserver, err := plan.APIServer.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		cd.APIServerConfig = apiserver.(*v1alpha1.APIServerConfig)

	}

	md := out.MachineConfig
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
			return &v1alpha1.Config{}, err
		}
		dev.(*v1alpha1.Device).DeviceInterface = netInterface
		md.MachineNetwork.NetworkInterfaces = append(md.MachineNetwork.NetworkInterfaces, dev.(*v1alpha1.Device))
	}
	for _, ns := range plan.Nameservers {
		md.MachineNetwork.NameServers = append(md.MachineNetwork.NameServers, ns.Value)
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
		InstallDisk:       plan.InstallDisk.Value,
		InstallImage:      plan.TalosImage.Value,
		InstallBootloader: true,
	}
	if plan.KernelArgs != nil {
		md.MachineInstall.InstallExtraKernelArgs = []string{}
		for _, arg := range plan.KernelArgs {
			md.MachineInstall.InstallExtraKernelArgs = append(md.MachineInstall.InstallExtraKernelArgs, arg.Value)
		}
	}

	if plan.MachineControlPlane != nil {
		mcp, err := plan.MachineControlPlane.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		md.MachineControlPlane = mcp.(*v1alpha1.MachineControlPlaneConfig)
	}

	for _, pod := range plan.Pod {
		var talosPod v1alpha1.Unstructured

		if err = yaml.Unmarshal([]byte(pod.Value), &talosPod); err != nil {
			return
		}

		md.MachinePods = append(md.MachinePods, talosPod)
	}

	md.MachineEnv = map[string]string{}
	for name, value := range plan.Env {
		md.MachineEnv[name] = value.Value
	}

	for _, planFile := range plan.Files {
		file, err := planFile.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
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

	if plan.Proxy != nil {
		proxy, err := plan.Proxy.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		cd.ProxyConfig = proxy.(*v1alpha1.ProxyConfig)
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

	for _, manifestURL := range plan.ExtraManifests {
		cd.ExtraManifests = append(cd.ExtraManifests, manifestURL.Value)
	}

	for _, planManifest := range plan.InlineManifests {
		manifest, err := planManifest.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		cd.ClusterInlineManifests = append(cd.ClusterInlineManifests, manifest.(v1alpha1.ClusterInlineManifest))
	}

	return
}

func (t talosControlNodeResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return talosControlNodeResource{
		provider: provider,
	}, diags
}

type talosControlNodeResource struct {
	provider provider
}

func (r talosControlNodeResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var (
		plan talosControlNodeResourceData
	)

	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos control node's Create method has been called without the provider being configured. This is a provider bug.")
		return
	}

	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := &plan
	config, errDesc, err := applyConfig(ctx, &p, configData{
		Bootstrap:   plan.Bootstrap.Value,
		ConfigIP:    plan.ConfigIP.Value,
		CreateNode:  true,
		Mode:        machine.ApplyConfigurationRequest_REBOOT,
		BaseConfig:  plan.BaseConfig.Value,
		MachineType: machinetype.TypeControlPlane,
		Network:     plan.DHCPNetworkCidr.Value,
		MAC:         plan.Macaddr.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError(errDesc, err.Error())
		return
	}

	plan.Patch = types.String{Value: config}

	plan.ID = types.String{Value: string(plan.Name.Value)}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r talosControlNodeResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var (
		state talosControlNodeResourceData
	)

	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos control node's Read method has been called without the provider being configured. This is a provider bug.")
		return
	}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Error getting plan state.", "")
		return
	}

	if !r.provider.skipread {
		conf, errDesc, err := readConfig(ctx, &state, readData{
			ConfigIP:   state.ConfigIP.Value,
			BaseConfig: state.BaseConfig.Value,
		})
		if err != nil {
			resp.Diagnostics.AddError(errDesc, err.Error())
			return
		}

		state.ReadInto(conf)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r talosControlNodeResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var (
		state talosControlNodeResourceData
	)

	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos control node's Update method has been called without the provider being configured. This is a provider bug.")
		return
	}

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := &state
	config, errDesc, err := applyConfig(ctx, &p, configData{
		Bootstrap:   false,
		ConfigIP:    state.ConfigIP.Value,
		Mode:        machine.ApplyConfigurationRequest_AUTO,
		BaseConfig:  state.BaseConfig.Value,
		MachineType: machinetype.TypeControlPlane,
		Network:     state.DHCPNetworkCidr.Value,
		MAC:         state.Macaddr.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError(errDesc, err.Error())
		return
	}

	state.Patch = types.String{Value: config}

	if !r.provider.skipread {
		talosConf, errDesc, err := readConfig(ctx, &state, readData{
			ConfigIP:   state.ConfigIP.Value,
			BaseConfig: state.BaseConfig.Value,
		})
		if err != nil {
			resp.Diagnostics.AddError(errDesc, err.Error())
			return
		}
		state.ReadInto(talosConf)
	}

	state.ID = types.String{Value: string(state.Name.Value)}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r talosControlNodeResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var (
		state talosControlNodeResourceData
	)

	if !r.provider.configured {
		resp.Diagnostics.AddError("Provider not configured.", "The Talos control node's Delete method has been called without the provider being configured. This is a provider bug.")
		return
	}

	if r.provider.skipdelete {
		resp.Diagnostics.AddError("skipdelete set", "skipdelete set")
		return
	}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := net.JoinHostPort(state.ConfigIP.Value, strconv.Itoa(talosPort))

	input := generate.Input{}
	if err := json.Unmarshal([]byte(state.BaseConfig.Value), &input); err != nil {
		resp.Diagnostics.AddError("error while unmarshalling Talos node bae configuration package", err.Error())
		return
	}

	conn, err := secureConn(ctx, input, host)
	if err != nil {
		resp.Diagnostics.AddError("error while attempting to connect to Talos API endpoint", err.Error())
		return
	}
	defer conn.Close()

	client := machine.NewMachineServiceClient(conn)
	_, err = client.Reset(ctx, &machine.ResetRequest{
		Graceful: false,
		Reboot:   true,
	})
	if err != nil {
		resp.Diagnostics.AddError("error while attempting to connect to reset maachine", err.Error())
		return
	}

	// Need to give the system enough time to perform the reset.
	// TODO figure out how to determine a machine is down. Probably by periodically pinging the machine's
	// config IP until a response is no longer recieved.
	time.Sleep(60 * time.Second)
}

func (r talosControlNodeResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

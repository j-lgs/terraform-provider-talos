package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"terraform-provider-talos/talos/datatypes"
	"time"

	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/talos-systems/talos/pkg/machinery/api/machine"

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
		MarkdownDescription: "Represents a Talos controlplane node.",
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
				// ValidateFunc: validateDomain,
				// ForceNew: true,
				// TODO validate and fix forcenew
			},
			"provision_ip": {
				Type:        types.StringType,
				Description: "IP address of the machine to be provisioned.",
				Required:    true,
				// TODO validate and forcenew
				// ForceNew: false
				// doesn't matter if changed after initial creation.
			},
			// --- MachineConfig.
			// See https://www.talos.dev/v1.0/reference/configuration/#machineconfig for full spec.

			"config": {
				Required:    true,
				Description: datatypes.TalosConfigSchema.MarkdownDescription,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.TalosConfigSchema.Attributes),
			},

			// ----- MachineConfig End
			// ----- ClusterConfig Start

			// ----- ClusterConfig End
			// ----- Resource Cluster bootstrap configuration
			"bootstrap": {
				Type:     types.BoolType,
				Required: true,
			},
			"configure_ip": {
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

var (
	talosControlNodeResourceDataExample = &talosControlNodeResourceData{
		Name:        datatypes.Wraps("test-node"),
		TalosConfig: *datatypes.TalosConfigExample,
	}
)

type talosControlNodeResourceData struct {
	Name types.String `tfsdk:"name"`

	datatypes.TalosConfig `tfsdk:"config"`

	Bootstrap   types.Bool   `tfsdk:"bootstrap"`
	ProvisionIP types.String `tfsdk:"provision_ip"`
	ConfigIP    types.String `tfsdk:"configure_ip"`
	BaseConfig  types.String `tfsdk:"base_config"`
	Patch       types.String `tfsdk:"patch"`
	ID          types.String `tfsdk:"id"`
}

func (plan *talosControlNodeResourceData) Generate() (err error) {
	input := generate.Input{}
	if err := json.Unmarshal([]byte(plan.BaseConfig.Value), &input); err != nil {
		return fmt.Errorf("unable to marshal node's base_config data into it's generate.Input struct: %w", err)
	}

	// Generate wireguard keys.
	for _, device := range plan.Network.Devices {
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

	// TODO derive these from talos machinery

	plan.ControlPlane = &datatypes.ControlPlaneConfig{
		Endpoint:           types.String{Value: input.GetControlPlaneEndpoint()},
		LocalAPIServerPort: types.Int64{Null: true},
	}

	plan.ControllerManager = &datatypes.ControllerManagerConfig{
		Image: types.String{Value: (&v1alpha1.SchedulerConfig{}).Image()},
	}

	plan.CoreDNS = &datatypes.CoreDNS{
		Image: types.String{Value: (&v1alpha1.CoreDNS{}).Image()},
	}

	plan.AllowSchedulingOnMasters = types.Bool{Value: input.AllowSchedulingOnMasters}

	plan.Kubelet = &datatypes.KubeletConfig{
		Image: types.String{Value: (&v1alpha1.KubeletConfig{}).Image()},
	}

	plan.Proxy = &datatypes.ProxyConfig{
		Image: types.String{Value: (&v1alpha1.ProxyConfig{}).Image()},
	}

	plan.Scheduler = &datatypes.SchedulerConfig{
		Image: types.String{Value: (&v1alpha1.SchedulerConfig{}).Image()},
	}

	plan.APIServer = &datatypes.APIServerConfig{
		Image:      types.String{Value: (&v1alpha1.APIServerConfig{}).Image()},
		DisablePSP: types.Bool{Value: bool(true)},
	}

	for _, san := range input.GetAPIServerSANs() {
		plan.APIServer.CertSANS = append(plan.APIServer.CertSANS, types.String{Value: san})
	}

	plan.Install.Image = types.String{Value: input.InstallImage}
	if input.InstallImage == "" {
		plan.Install.Image = types.String{Value: generate.DefaultGenOptions().InstallDisk}
	}

	plan.Discovery = &datatypes.ClusterDiscoveryConfig{
		Enabled: types.Bool{Value: input.DiscoveryEnabled},
	}

	plan.Etcd = &datatypes.EtcdConfig{
		Image: types.String{Value: (&v1alpha1.EtcdConfig{}).Image()},
		CaCrt: types.String{Value: string(input.Certs.Etcd.Crt)},
		CaKey: types.String{Value: string(input.Certs.Etcd.Key)},
	}

	return
}

func (plan *talosControlNodeResourceData) ReadInto(in *v1alpha1.Config) (err error) {
	if in == nil {
		return
	}

	funcs := []datatypes.ConfigToPlanFunc{
		datatypes.TalosKubelet{KubeletConfig: in.MachineConfig.MachineKubelet},
		datatypes.TalosProxyConfig{ProxyConfig: in.ClusterConfig.ProxyConfig},
		datatypes.TalosRegistriesConfig{RegistriesConfig: &in.MachineConfig.MachineRegistries},
		datatypes.TalosMCPConfig{MachineControlPlaneConfig: in.MachineConfig.MachineControlPlane},
		datatypes.TalosSystemDiskEncryptionConfig{SystemDiskEncryptionConfig: in.MachineConfig.MachineSystemDiskEncryption},
		datatypes.TalosInstallConfig{InstallConfig: in.MachineConfig.MachineInstall},
		datatypes.TalosMachineDisk{MachineDisks: in.MachineConfig.MachineDisks},
		datatypes.TalosNetworkConfig{NetworkConfig: in.MachineConfig.MachineNetwork},
		datatypes.TalosAPIServerConfig{APIServerConfig: in.ClusterConfig.APIServerConfig},
		datatypes.TalosControlPlaneConfig{ControlPlaneConfig: in.ClusterConfig.ControlPlane},
		datatypes.TalosMachineSysfs(in.MachineConfig.MachineSysfs),
		datatypes.TalosMachineSysctls(in.MachineConfig.MachineSysctls),
		datatypes.TalosFiles{Files: in.MachineConfig.MachineFiles},
		datatypes.TalosTimeConfig{TimeConfig: in.MachineConfig.MachineTime},
		datatypes.TalosKernelConfig{KernelConfig: in.MachineConfig.MachineKernel},
		datatypes.TalosLoggingConfig{LoggingConfig: in.MachineConfig.MachineLogging},
		datatypes.TalosSchedulerConfig{SchedulerConfig: in.ClusterConfig.SchedulerConfig},
		datatypes.TalosClusterDiscoveryConfig{ClusterDiscoveryConfig: &in.ClusterConfig.ClusterDiscoveryConfig},
		datatypes.TalosEtcdConfig{EtcdConfig: in.ClusterConfig.EtcdConfig},
		datatypes.TalosCoreDNS{CoreDNS: in.ClusterConfig.CoreDNSConfig},
		datatypes.TalosAdminKubeconfigConfig{AdminKubeconfigConfig: in.ClusterConfig.AdminKubeconfigConfig},
		datatypes.TalosControllerManagerConfig{ControllerManagerConfig: in.ClusterConfig.ControllerManagerConfig},
		datatypes.TalosClusterInlineManifests{ClusterInlineManifests: in.ClusterConfig.ClusterInlineManifests},
		datatypes.TalosMachineEnv(in.MachineConfig.MachineEnv),
		datatypes.TalosExtraManifestHeaders(in.ClusterConfig.ExtraManifestHeaders),
		datatypes.TalosMachineUdev{UdevConfig: in.MachineConfig.MachineUdev},
		datatypes.TalosMachineCertSANs(in.MachineConfig.MachineCertSANs),
		datatypes.TalosClusterExtraManifests(in.ClusterConfig.ExtraManifests),
		datatypes.TalosMachinePods(in.MachineConfig.MachinePods),
		datatypes.TalosExternalCloudProvider{ExternalCloudProviderConfig: in.ClusterConfig.ExternalCloudProviderConfig},
	}

	readFuncs := []datatypes.ConfigReadFunc{}
	readFuncs = datatypes.AppendReadFunc(readFuncs, funcs...)
	if plan.TalosConfig, err = datatypes.ApplyReadFunc(&plan.TalosConfig, readFuncs); err != nil {
		return fmt.Errorf("error applying read functions: %w", err)
	}

	if in.ClusterConfig.AllowSchedulingOnMasters {
		plan.AllowSchedulingOnMasters = types.Bool{Value: in.ClusterConfig.AllowSchedulingOnMasters}
		plan.AllowSchedulingOnMasters.Null = false
	}

	return nil
}

func (plan *talosControlNodeResourceData) TalosData(in *v1alpha1.Config) (out *v1alpha1.Config, err error) {
	out = &v1alpha1.Config{}
	in.DeepCopyInto(out)

	clusterFuncs := []datatypes.ConfigDataFunc{}
	funcs := []datatypes.PlanToDataFunc{
		plan.Kubelet,
		plan.Proxy,
		plan.Registry,
		plan.MachineControlPlane,
		plan.Encryption,
		plan.Install,
		plan.Network,
		plan.APIServer,
		plan.ControlPlane,
		plan.Sysfs,
		plan.Sysctls,
		plan.Env,
		plan.Time,
		plan.Logging,
		plan.Kernel,
		plan.ControllerManager,
		plan.Scheduler,
		plan.Discovery,
		plan.Etcd,
		plan.CoreDNS,
		plan.AdminKubeConfig,
		plan.ExtraManifestHeaders,
		plan.CertSANS,
		plan.Udev,
		plan.ExtraManifests,
		plan.Pod,
		plan.ExternalCloudProvider,
	}
	for _, file := range plan.Files {
		funcs = append(funcs, any(file).(datatypes.PlanToDataFunc))
	}
	for _, manifest := range plan.InlineManifests {
		funcs = append(funcs, any(manifest).(datatypes.InlineManifest))
	}

	clusterFuncs = datatypes.AppendDataFunc(clusterFuncs, funcs...)
	if err := datatypes.ApplyDataFunc(out, clusterFuncs); err != nil {
		return nil, err
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
		ProvisionIP: plan.ProvisionIP.Value,
		CreateNode:  true,
		Mode:        machine.ApplyConfigurationRequest_REBOOT,
		BaseConfig:  plan.BaseConfig.Value,
		MachineType: machinetype.TypeControlPlane,
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

		if err = state.ReadInto(conf); err != nil {
			resp.Diagnostics.AddError("Error reading talos configuration.", err.Error())
			return
		}
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

	diags := req.Plan.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := &state
	config, errDesc, err := applyConfig(ctx, &p, configData{
		Bootstrap:   false,
		ProvisionIP: state.ProvisionIP.Value,
		ConfigIP:    state.ConfigIP.Value,
		Mode:        machine.ApplyConfigurationRequest_AUTO,
		BaseConfig:  state.BaseConfig.Value,
		MachineType: machinetype.TypeControlPlane,
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

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.provider.skipdelete {
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
		resp.Diagnostics.AddError("error while attempting to connect to reset machine", err.Error())
		return
	}

	// The testing environment has issues regarding reboots
	// Here we will manually send a command to the qemu socket to forcefully reset the machine.
	if val, set := os.LookupEnv("TF_ACC"); set {
		// This approach is not ideal is it might take much more or much less time for a talos host to
		// reset. Ideally there would be an insecure endpoint that can be checked to determine if a host
		// is up. Likely it would return 200 if up. This is handy too as it can help the provider determine
		// whether the host's networking stack is up.
		time.Sleep(50 * time.Second)

		// Require more time if inside a Github Action
		if _, set := os.LookupEnv("GITHUB_ACTIONS"); set {
			time.Sleep(60 * time.Second)
		}

		b, err := strconv.ParseBool(val)
		if err != nil {
			resp.Diagnostics.AddError("environment parse error",
				"error parsing boolean value for TF_ACC:"+err.Error())
			return
		}

		if !b {
			return
		}

		hostname := state.Network.Hostname.Value
		conn, err := net.Dial("unix", "/tmp/qmp/vm-"+hostname+".sock")
		if err != nil {
			resp.Diagnostics.AddError("VM socket connect error",
				"Issue connecting to VM socket at /tmp/qmp/vm-"+hostname+".sock: "+err.Error())
			return
		}
		defer conn.Close()

		buf := make([]byte, 256)
		if n, err := conn.Read(buf); n <= 0 || err != nil {
			resp.Diagnostics.AddError("VM socket read error",
				"got "+strconv.Itoa(n)+"bytes. error: "+err.Error())
			return
		}

		conn.Write([]byte(`{"execute": "qmp_capabilities"}`))
		if n, err := conn.Read(buf); n <= 0 || err != nil {
			resp.Diagnostics.AddError("VM socket read error",
				"got "+strconv.Itoa(n)+"bytes. error: "+err.Error())
			return
		}

		conn.Write([]byte(`{"execute": "system_reset"}`))
		if n, err := conn.Read(buf); n <= 0 || err != nil {
			resp.Diagnostics.AddError("VM socket read error",
				"got "+strconv.Itoa(n)+"bytes. error: "+err.Error())
			return
		}
	}
}

func (r talosControlNodeResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

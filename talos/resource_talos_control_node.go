package talos

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"terraform-provider-talos/talos/datatypes"
	"time"

	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"

	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

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

			"install": {
				Required:    true,
				Description: datatypes.InstallSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.InstallSchema.Attributes),
			},
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
				Description: datatypes.ControlPlaneConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.ControlPlaneConfigSchema.Attributes),
			},
			"kubelet": {
				Optional:    true,
				Description: datatypes.KubeletConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.KubeletConfigSchema.Attributes),
			},
			"pods": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
				// TODO validation
				Description: "Used to provide static pod definitions to be run by the kubelet directly bypassing the kube-apiserver.",
			},
			"network": {
				Required:    true,
				Description: datatypes.NetworkConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.NetworkConfigSchema.Attributes),
			},
			"disks": {
				Optional:    true,
				Description: datatypes.MachineDiskSchema.MarkdownDescription,
				Attributes:  tfsdk.ListNestedAttributes(datatypes.MachineDiskSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			},
			"files": {
				Optional:    true,
				Description: datatypes.FileSchema.Description,
				Attributes:  tfsdk.ListNestedAttributes(datatypes.FileSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			},
			"env": {
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Description: "Allows for the addition of environment variables. All environment variables are set on PID 1 in addition to every service.",
			},
			"time": {
				Description: datatypes.TimeConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.TimeConfigSchema.Attributes),
				Optional:    true,
			},
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
				Description: datatypes.RegistrySchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.RegistrySchema.Attributes),
			},
			"encryption": {
				Optional:    true,
				Description: datatypes.EncryptionSchema.MarkdownDescription,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.EncryptionSchema.Attributes),
			},
			// features not implemented
			"udev": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Description: "Configures the udev system.",
				Optional:    true,
			},
			"logging": {
				Description: datatypes.LoggingConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.LoggingConfigSchema.Attributes),
				Optional:    true,
			},
			"kernel": {
				Description: datatypes.KernelConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.KernelConfigSchema.Attributes),
				Optional:    true,
			},
			// ----- MachineConfig End
			// ----- ClusterConfig Start

			"control_plane_config": {
				Optional:    true,
				Description: datatypes.MachineControlPlaneSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.MachineControlPlaneSchema.Attributes),
			},

			// clustername already filled
			// cluster_network not implemented
			"apiserver": {
				Optional:    true,
				Description: datatypes.APIServerConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.APIServerConfigSchema.Attributes),
			},
			"controller_manager": {
				Description: datatypes.ControllerManagerConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.ControllerManagerConfigSchema.Attributes),
				Optional:    true,
			},
			"proxy": {
				Optional:    true,
				Description: datatypes.ProxyConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.ProxyConfigSchema.Attributes),
			},
			"scheduler": {
				Description: datatypes.SchedulerConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.SchedulerConfigSchema.Attributes),
				Optional:    true,
			},
			"discovery": {
				Description: datatypes.ClusterDiscoveryConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.ClusterDiscoveryConfigSchema.Attributes),
				Optional:    true,
			},
			"etcd": {
				Description: datatypes.EtcdConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.EtcdConfigSchema.Attributes),
				Optional:    true,
			},
			"coredns": {
				Description: datatypes.CoreDNSConfigSchema.MarkdownDescription,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.CoreDNSConfigSchema.Attributes),
				Optional:    true,
			},
			"external_cloud_provider": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Description: ".",
				Optional:    true,
			},
			"extra_manifests": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Description: "A list of urls that point to additional manifests. These will get automatically deployed as part of the bootstrap.",
				Optional:    true,
			},
			// TODO Add verification function confirming it's a correct manifest that can be downloaded.
			"inline_manifests": {
				Optional:    true,
				Description: datatypes.InlineManifestSchema.Description,
				Attributes:  tfsdk.ListNestedAttributes(datatypes.InlineManifestSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
			},
			"admin_kube_config": {
				Description: datatypes.AdminKubeconfigConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.AdminKubeconfigConfigSchema.Attributes),
				Optional:    true,
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
	CertSANsExample = []datatypes.MachineCertSAN{
		{Value: datatypes.Wraps(datatypes.MachineCertSANsExample[0])},
	}

	PodsExample = []datatypes.MachinePod{
		{Value: datatypes.Wraps(datatypes.MachinePodsStringExample)},
	}

	UdevExample = []datatypes.MachineUdev{
		{Value: datatypes.Wraps(datatypes.UdevExample[0])},
	}

	ExtraManifestExample = []datatypes.ClusterExtraManifest{
		{Value: datatypes.Wraps(datatypes.ExtraManifestExample[0])},
		{Value: datatypes.Wraps(datatypes.ExtraManifestExample[1])},
	}

	talosControlNodeResourceDataExample = &talosControlNodeResourceData{
		Name:         datatypes.Wraps("test-node"),
		Install:      datatypes.InstallExample,
		CertSANS:     CertSANsExample,
		ControlPlane: datatypes.ControlPlaneConfigExample,
		Kubelet:      datatypes.KubeletExample,
		Pod:          PodsExample,
		Network:      datatypes.NetworkConfigExample,
		Files: []datatypes.File{
			datatypes.FileExample,
		},
		Env: map[string]types.String{
			"GRPC_GO_LOG_VERBOSITY_LEVEL": datatypes.Wraps("99"),
			"GRPC_GO_LOG_SEVERITY_LEVEL":  datatypes.Wraps("info"),
			"https_proxy":                 datatypes.Wraps("http://DOMAIN\\USERNAME:PASSWORD@SERVER:PORT/"),
		},
		Sysctls: map[string]types.String{
			"kernel.domainname":   datatypes.Wraps("talos.dev"),
			"net.ipv4.ip_forward": datatypes.Wraps("0"),
		},
		Sysfs: map[string]types.String{
			"devices.system.cpu.cpu0.cpufreq.scaling_governor": datatypes.Wraps("performance"),
		},
		Disks: []datatypes.MachineDiskData{
			*datatypes.MachineDiskExample,
		},
		Time:                datatypes.TimeConfigExample,
		Logging:             datatypes.LoggingConfigExample,
		Kernel:              datatypes.KernelConfigExample,
		Encryption:          datatypes.EncryptionDataExample,
		Registry:            datatypes.RegistryExample,
		Udev:                UdevExample,
		MachineControlPlane: datatypes.MachineControlPlaneExample,
		APIServer:           datatypes.APIServerExample,
		ControllerManager:   datatypes.ControllerManagerExample,
		Proxy:               datatypes.ProxyConfigExample,
		Scheduler:           datatypes.SchedulerExample,
		Discovery:           datatypes.ClusterDiscoveryConfigExample,
		Etcd:                datatypes.EtcdConfigExample,
		CoreDNS:             datatypes.CoreDNSExample,
		ExtraManifests:      ExtraManifestExample,
		InlineManifests: []datatypes.InlineManifest{
			datatypes.InlineManifestExample,
		},
		AdminKubeConfig:          datatypes.AdminKubeconfigConfigExample,
		AllowSchedulingOnMasters: datatypes.Wrapb(datatypes.AllowSchedulingOnMastersExample),
	}
)

type talosControlNodeResourceData struct {
	Name                     types.String                       `tfsdk:"name"`
	Install                  *datatypes.InstallConfig           `tfsdk:"install"`
	CertSANS                 []datatypes.MachineCertSAN         `tfsdk:"cert_sans"`
	ControlPlane             *datatypes.ControlPlaneConfig      `tfsdk:"control_plane"`
	Kubelet                  *datatypes.KubeletConfig           `tfsdk:"kubelet"`
	Pod                      []datatypes.MachinePod             `tfsdk:"pods"`
	Network                  *datatypes.NetworkConfig           `tfsdk:"network"`
	Files                    []datatypes.File                   `tfsdk:"files"`
	Env                      datatypes.MachineEnv               `tfsdk:"env"`
	Time                     *datatypes.TimeConfig              `tfsdk:"time"`
	Logging                  *datatypes.LoggingConfig           `tfsdk:"logging"`
	Kernel                   *datatypes.KernelConfig            `tfsdk:"kernel"`
	Sysctls                  datatypes.MachineSysctls           `tfsdk:"sysctls"`
	Sysfs                    datatypes.MachineSysfs             `tfsdk:"sysfs"`
	Registry                 *datatypes.Registry                `tfsdk:"registry"`
	Disks                    []datatypes.MachineDiskData        `tfsdk:"disks"`
	Encryption               *datatypes.EncryptionData          `tfsdk:"encryption"`
	Udev                     []datatypes.MachineUdev            `tfsdk:"udev"`
	MachineControlPlane      *datatypes.MachineControlPlane     `tfsdk:"control_plane_config"`
	APIServer                *datatypes.APIServerConfig         `tfsdk:"apiserver"`
	ControllerManager        *datatypes.ControllerManagerConfig `tfsdk:"controller_manager"`
	Proxy                    *datatypes.ProxyConfig             `tfsdk:"proxy"`
	Scheduler                *datatypes.SchedulerConfig         `tfsdk:"scheduler"`
	Discovery                *datatypes.ClusterDiscoveryConfig  `tfsdk:"discovery"`
	Etcd                     *datatypes.EtcdConfig              `tfsdk:"etcd"`
	CoreDNS                  *datatypes.CoreDNS                 `tfsdk:"coredns"`
	ExtraManifests           []datatypes.ClusterExtraManifest   `tfsdk:"extra_manifests"`
	InlineManifests          []datatypes.InlineManifest         `tfsdk:"inline_manifests"`
	AdminKubeConfig          *datatypes.AdminKubeconfigConfig   `tfsdk:"admin_kube_config"`
	AllowSchedulingOnMasters types.Bool                         `tfsdk:"allow_scheduling_on_masters"`
	Bootstrap                types.Bool                         `tfsdk:"bootstrap"`

	ProvisionIP types.String `tfsdk:"provision_ip"`
	ConfigIP    types.String `tfsdk:"configure_ip"`
	BaseConfig  types.String `tfsdk:"base_config"`
	Patch       types.String `tfsdk:"patch"`
	ID          types.String `tfsdk:"id"`
}

func (plan *talosControlNodeResourceData) Generate() (err error) {
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

	return
}

func (plan *talosControlNodeResourceData) ReadInto(in *v1alpha1.Config) (err error) {
	if in == nil {
		return
	}

	plan.Network = &datatypes.NetworkConfig{}
	plan.Network.Nameservers = []types.String{}
	for _, ns := range in.MachineConfig.MachineNetwork.NameServers {
		plan.Network.Nameservers = append(plan.Network.Nameservers, types.String{Value: ns})
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
	}
	//funcs = datatypes.AppendDataFuncs(funcs, datatypes.ToSliceOfAny(plan.Files))
	funcs = datatypes.AppendDataFuncs(funcs, datatypes.ToSliceOfAny(plan.CertSANS))
	funcs = datatypes.AppendDataFuncs(funcs, datatypes.ToSliceOfAny(plan.Udev))
	funcs = datatypes.AppendDataFuncs(funcs, datatypes.ToSliceOfAny(plan.ExtraManifests))
	funcs = datatypes.AppendDataFuncs(funcs, datatypes.ToSliceOfAny(plan.InlineManifests))
	funcs = datatypes.AppendDataFuncs(funcs, datatypes.ToSliceOfAny(plan.Pod))

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

	/*
			config, err := r.provider.client.GetConfig()
			if err != nil {
				resp.Diagnostics.AddError(errDesc, err.Error())
				return
			}

			err := r.provider.createNode(machinetype.TypeControlPlane, plan.ProvisionIP.Value)
			if err != nil {
				resp.Diagnostics.AddError(errDesc, err.Error())
				return
			}

			if plan.Bootstrap.Value {
				err := r.provider.client.Bootstrap(plan.ConfigIP.Value)
			        if err != nil {
		  		  resp.Diagnostics.AddError(errDesc, err.Error())
				  return
			        }
			}
	*/
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
		/*
			conf, errDesc, err := readConfig(ctx, &state, readData{
				ConfigIP:   state.ConfigIP.Value,
				BaseConfig: state.BaseConfig.Value,
			})
			if err != nil {
				resp.Diagnostics.AddError(errDesc, err.Error())
				return
			}
			conf = nil
		*/
		state.ReadInto(nil)
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

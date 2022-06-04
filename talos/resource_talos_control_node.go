package talos

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"strconv"
	"terraform-provider-talos/talos/datatypes"
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

			// logging not implemented
			// kernel not implemented
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

			// controller manager not implemented
			"proxy": {
				Optional:    true,
				Description: datatypes.ProxyConfigSchema.Description,
				Attributes:  tfsdk.SingleNestedAttributes(datatypes.ProxyConfigSchema.Attributes),
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
				Description: datatypes.InlineManifestSchema.Description,
				Attributes:  tfsdk.ListNestedAttributes(datatypes.InlineManifestSchema.Attributes, tfsdk.ListNestedAttributesOptions{}),
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
	talosControlNodeResourceDataExample = &talosControlNodeResourceData{
		Name:         datatypes.Wraps("test-node"),
		Install:      datatypes.InstallExample,
		CertSANS:     datatypes.Wrapsl(datatypes.MachineCertSANsExample...),
		ControlPlane: datatypes.ControlPlaneConfigExample,
		Kubelet:      datatypes.KubeletExample,
		Pod:          datatypes.Wrapsl(datatypes.MachinePodsStringExample),
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
		Encryption:          datatypes.EncryptionDataExample,
		Registry:            datatypes.RegistryExample,
		Udev:                datatypes.Wrapsl(datatypes.UdevExample...),
		MachineControlPlane: datatypes.MachineControlPlaneExample,
		APIServer:           datatypes.APIServerExample,
		Proxy:               datatypes.ProxyConfigExample,
		ExtraManifests:      datatypes.Wrapsl(datatypes.ExtraManifestExample...),
		InlineManifests: []datatypes.InlineManifest{
			datatypes.InlineManifestExample,
		},
		AllowSchedulingOnMasters: datatypes.Wrapb(datatypes.AllowSchedulingOnMastersExample),
	}
)

type talosControlNodeResourceData struct {
	Name                     types.String                   `tfsdk:"name"`
	Install                  *datatypes.InstallConfig       `tfsdk:"install"`
	CertSANS                 []types.String                 `tfsdk:"cert_sans"`
	ControlPlane             *datatypes.ControlPlaneConfig  `tfsdk:"control_plane"`
	Kubelet                  *datatypes.KubeletConfig       `tfsdk:"kubelet"`
	Pod                      []types.String                 `tfsdk:"pods"`
	Network                  *datatypes.NetworkConfig       `tfsdk:"network"`
	Files                    []datatypes.File               `tfsdk:"files"`
	Env                      map[string]types.String        `tfsdk:"env"`
	Sysctls                  map[string]types.String        `tfsdk:"sysctls"`
	Sysfs                    map[string]types.String        `tfsdk:"sysfs"`
	Registry                 *datatypes.Registry            `tfsdk:"registry"`
	Disks                    []datatypes.MachineDiskData    `tfsdk:"disks"`
	Encryption               *datatypes.EncryptionData      `tfsdk:"encryption"`
	Udev                     []types.String                 `tfsdk:"udev"`
	MachineControlPlane      *datatypes.MachineControlPlane `tfsdk:"control_plane_config"`
	APIServer                *datatypes.APIServerConfig     `tfsdk:"apiserver"`
	Proxy                    *datatypes.ProxyConfig         `tfsdk:"proxy"`
	ExtraManifests           []types.String                 `tfsdk:"extra_manifests"`
	InlineManifests          []datatypes.InlineManifest     `tfsdk:"inline_manifests"`
	AllowSchedulingOnMasters types.Bool                     `tfsdk:"allow_scheduling_on_masters"`
	Bootstrap                types.Bool                     `tfsdk:"bootstrap"`

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
	cd := out.ClusterConfig
	if plan.ControlPlane != nil {
		// Refrain from setting this for now because it's set in the cluster config resource
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
	if plan.Network != nil {
		net, err := plan.Network.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		md.MachineNetwork = net.(*v1alpha1.NetworkConfig)
	}

	if plan.Install != nil {
		install, err := plan.Install.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		md.MachineInstall = install.(*v1alpha1.InstallConfig)
	}
	if plan.Encryption != nil {
		encryption, err := plan.Encryption.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		md.MachineSystemDiskEncryption = encryption.(*v1alpha1.SystemDiskEncryptionConfig)
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

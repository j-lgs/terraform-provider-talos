package talos

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opencontainers/runtime-spec/specs-go"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func generateCommonConfig(d *schema.ResourceData, config *v1alpha1.Config) diag.Diagnostics {
	mc := config.MachineConfig
	cc := config.ClusterConfig

	// Install configuration
	mc.MachineInstall.InstallDisk = d.Get("install_disk").(string)
	mc.MachineInstall.InstallImage = d.Get("talos_image").(string)
	mc.MachineInstall.InstallExtraKernelArgs = GetTypeList[string](d.Get("kernel_args"))

	// Cluster configuration
	*cc.ProxyConfig = GetTypeMapWrap(d.Get("cluster_proxy_args"), func(v map[string]string) v1alpha1.ProxyConfig {
		return v1alpha1.ProxyConfig{
			ExtraArgsConfig: v,
		}
	})

	*cc.APIServerConfig = GetTypeMapWrap(d.Get("cluster_apiserver_args"), func(v map[string]string) v1alpha1.APIServerConfig {
		return v1alpha1.APIServerConfig{
			ExtraArgsConfig: v,
		}
	})

	// Network configuration
	mc.MachineNetwork.NetworkHostname = d.Get("name").(string)
	mc.MachineNetwork.NameServers = GetTypeList[string](d.Get("nameservers"))

	ifs, err := ExpandDeviceList(d.Get("interface").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}
	mc.MachineNetwork.NetworkInterfaces = ifs

	for _, mount := range d.Get("kubelet_extra_mount").([]interface{}) {
		m := mount.(map[string]interface{})

		mountOptions := []string{}
		for _, option := range m["options"].([]interface{}) {
			mountOptions = append(mountOptions, option.(string))
		}

		mc.MachineKubelet.KubeletExtraMounts = append(mc.MachineKubelet.KubeletExtraMounts, v1alpha1.ExtraMount{
			Mount: specs.Mount{
				Destination: m["destination"].(string),
				Type:        m["type"].(string),
				Source:      m["source"].(string),
				Options:     mountOptions,
			},
		})
	}
	/*
		mc.MachineRegistries.RegistryMirrors = GetTypeMapWrapEach(d.Get("registry_mirrors"), func(v string) *v1alpha1.RegistryMirrorConfig {
			return &v1alpha1.RegistryMirrorConfig{
				MirrorEndpoints: []string{v},
			}
		})
	*/
	mc.MachineKubelet.KubeletExtraArgs = GetTypeMap[string](d.Get("kubelet_extra_args"))
	mc.MachineSysctls = GetTypeMap[string](d.Get("sysctls"))

	return nil
}

// Errors for the ExpandDeviceList function
var (
	WireguardExtraFieldError = fmt.Errorf("There can only be one wireguard field in each network device.")
)

// ExpandsDeviceList extracts the values from the "interface" schema value shared between node resources and returns
// a list of Talos network devices. It needs to handle an optional "wireguard" field, max of one, and optional route fields.
// See https://www.talos.dev/v1.0/reference/configuration/#device for more information about the output spec.
func ExpandDeviceList(interfaces []interface{}) (devices []*v1alpha1.Device, err error) {
	for _, netInterface := range interfaces {
		n := netInterface.(map[string]interface{})
		dev := &v1alpha1.Device{}

		// The interface name
		dev.DeviceInterface = n["name"].(string)
		// Static IP addresses for the interface
		dev.DeviceAddresses = GetTypeList[string](n["addresses"])

		// Expand the list of interface maps that form's the "route" schema
		routes := []*v1alpha1.Route{}
		for _, resourceRoute := range n["route"].([]interface{}) {
			r := resourceRoute.(map[string]interface{})

			routes = append(routes, &v1alpha1.Route{
				RouteGateway: r["gateway"].(string),
				RouteNetwork: r["network"].(string),
				RouteMetric:  uint32(r["metric"].(int)), // optional
			})
		}
		if len(routes) > 0 {
			dev.DeviceRoutes = routes
		}

		// The interface's MTU
		dev.DeviceMTU = n["mtu"].(int)

		wg_ := n["wireguard"].([]interface{})
		// Ensure there can only be one instance of the wireguard interface, Return an error if there are more
		if len(wg_) > 1 {
			return nil, WireguardExtraFieldError
		}
		if len(wg_) == 1 {
			wg := wg_[0].(map[string]interface{})

			// Expand the list of interface maps that form's the "wireguard" schema's "peer" element
			peers := []*v1alpha1.DeviceWireguardPeer{}
			for _, peer := range wg["peer"].([]interface{}) {
				p := peer.(map[string]interface{})

				peers = append(peers, &v1alpha1.DeviceWireguardPeer{
					WireguardAllowedIPs:                  GetTypeList[string](p["allowed_ips"]),
					WireguardEndpoint:                    p["endpoint"].(string),
					WireguardPersistentKeepaliveInterval: time.Duration(p["persistent_keepalive_interval"].(int)) * time.Second,
					WireguardPublicKey:                   p["public_key"].(string),
				})
			}

			dev.DeviceWireguardConfig = &v1alpha1.DeviceWireguardConfig{
				WireguardPrivateKey: wg["private_key"].(string),
				WireguardPeers:      peers,
			}
		}

		devices = append(devices, dev)
	}

	return
}

/*
			// Workaround for the inability for terraform to perform set on an element in list resources.
			// Wants to set a full list.
			interfaces_ := interfaces
			interfaces_[i].(map[string]interface{})["wireguard"].([]interface{})[0].(map[string]interface{})["public_key"] = pub
			interfaces_[i].(map[string]interface{})["wireguard"].([]interface{})[0].(map[string]interface{})["private_key"] = priv

			if err := d.Set("interface", interfaces_); err != nil {
				return nil, diag.FromErr(err)
}
*/

type MachineWrap interface {
	v1alpha1.APIServerConfig | v1alpha1.ProxyConfig
}

type MachineWrapEach interface {
	*v1alpha1.RegistryMirrorConfig
}

type ListT interface {
	string | int
}

func GetTypeMapWrapEach[T MachineWrapEach](assignFrom interface{}, f func(interface{}) T) (assignTo map[string]T) {
	assignTo = map[string]T{}
	for k, v := range assignFrom.(map[string]interface{}) {
		assignTo[k] = f(v)
	}

	return
}

func GetTypeMapWrap[T MachineWrap](assignFrom interface{}, f func(map[string]string) T) (assignTo T) {
	temp := map[string]string{}
	for k, v := range assignFrom.(map[string]interface{}) {
		temp[k] = v.(string)
	}

	assignTo = f(temp)
	return
}

func GetTypeList[T ListT](assignFrom interface{}) (assignTo []T) {
	assignTo = make([]T, len(assignFrom.([]interface{})))
	if assignFrom == nil {
		return
	}
	for i, v := range assignFrom.([]interface{}) {
		assignTo[i] = v.(T)
	}

	return
}

// Assign to a string map from a map of interfaces, indexed by strings, used to get values from TypeMaps
func GetTypeMap[T ListT](assignFrom interface{}) (assignTo map[string]T) {
	assignTo = map[string]T{}
	for k, v := range assignFrom.(map[string]interface{}) {
		assignTo[k] = v.(T)
	}
	return
}

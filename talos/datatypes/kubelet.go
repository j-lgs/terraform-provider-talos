package datatypes

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"gopkg.in/yaml.v3"
)

func (kubelet KubeletConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	funs := [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			talosKubelet := &v1alpha1.KubeletConfig{}
			if !kubelet.Image.Null {
				talosKubelet.KubeletImage = kubelet.Image.Value
			}
			if !kubelet.RegisterWithFQDN.Null {
				talosKubelet.KubeletRegisterWithFQDN = kubelet.RegisterWithFQDN.Value
			}
			if !kubelet.ExtraConfig.Null {
				var conf v1alpha1.Unstructured
				if err := yaml.Unmarshal([]byte(kubelet.ExtraConfig.Value), &conf); err != nil {
					return err
				}

				talosKubelet.KubeletExtraConfig = conf
			}
			for _, dns := range kubelet.ClusterDNS {
				talosKubelet.KubeletClusterDNS = append(talosKubelet.KubeletClusterDNS, dns.Value)
			}
			if len(kubelet.ExtraArgs) > 0 {
				talosKubelet.KubeletExtraArgs = map[string]string{}
			}
			for k, arg := range kubelet.ExtraArgs {
				talosKubelet.KubeletExtraArgs[k] = arg.Value
			}
			for _, mount := range kubelet.ExtraMounts {
				m, err := mount.Data()
				if err != nil {
					return err
				}
				talosKubelet.KubeletExtraMounts = append(talosKubelet.KubeletExtraMounts, m.(v1alpha1.ExtraMount))
			}
			if len(kubelet.NodeIPValidSubnets) > 0 {
				talosKubelet.KubeletNodeIP = v1alpha1.KubeletNodeIPConfig{}
				for _, subnet := range kubelet.NodeIPValidSubnets {
					talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets =
						append(talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets, subnet.Value)
				}
			}

			cfg.MachineConfig.MachineKubelet = talosKubelet

			return nil
		},
	}
	return funs
}

func (kubelet KubeletConfig) zero() bool {
	return mkString(kubelet.Image).zero() &&
		mkBool(kubelet.RegisterWithFQDN).zero() &&
		mkString(kubelet.ExtraConfig.Value).zero() &&
		len(kubelet.ClusterDNS) <= 0 &&
		len(kubelet.ExtraMounts) <= 0 &&
		len(kubelet.NodeIPValidSubnets) <= 0 &&
		len(kubelet.ExtraArgs) <= 0
}

// Data copies data from terraform state types to talos types.
func (kubelet KubeletConfig) Data() (interface{}, error) {
	talosKubelet := &v1alpha1.KubeletConfig{}
	if !kubelet.Image.Null {
		talosKubelet.KubeletImage = kubelet.Image.Value
	}
	if !kubelet.RegisterWithFQDN.Null {
		talosKubelet.KubeletRegisterWithFQDN = kubelet.RegisterWithFQDN.Value
	}
	if !kubelet.ExtraConfig.Null {
		var conf v1alpha1.Unstructured
		if err := yaml.Unmarshal([]byte(kubelet.ExtraConfig.Value), &conf); err != nil {
			return nil, nil
		}

		talosKubelet.KubeletExtraConfig = conf
	}
	for _, dns := range kubelet.ClusterDNS {
		talosKubelet.KubeletClusterDNS = append(talosKubelet.KubeletClusterDNS, dns.Value)
	}
	if len(kubelet.ExtraArgs) > 0 {
		talosKubelet.KubeletExtraArgs = map[string]string{}
	}
	for k, arg := range kubelet.ExtraArgs {
		talosKubelet.KubeletExtraArgs[k] = arg.Value
	}
	for _, mount := range kubelet.ExtraMounts {
		m, err := mount.Data()
		if err != nil {
			return nil, err
		}
		talosKubelet.KubeletExtraMounts = append(talosKubelet.KubeletExtraMounts, m.(v1alpha1.ExtraMount))
	}
	if len(kubelet.NodeIPValidSubnets) > 0 {
		talosKubelet.KubeletNodeIP = v1alpha1.KubeletNodeIPConfig{}
		for _, subnet := range kubelet.NodeIPValidSubnets {
			talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets =
				append(talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets, subnet.Value)
		}
	}
	return talosKubelet, nil
}

// Read copies data from talos types to terraform state types.
func (kubelet *KubeletConfig) Read(talosData interface{}) error {
	talosKubelet := talosData.(*v1alpha1.KubeletConfig)
	if talosKubelet.KubeletImage != "" {
		kubelet.Image = types.String{Value: talosKubelet.KubeletImage}
	}

	kubelet.RegisterWithFQDN = types.Bool{Value: talosKubelet.KubeletRegisterWithFQDN}

	if !reflect.DeepEqual(talosKubelet.KubeletExtraConfig.Object, map[string]interface{}{}) {
		bytes, err := yaml.Marshal(&talosKubelet.KubeletExtraConfig)
		if err != nil {
			return err
		}
		kubelet.ExtraConfig = types.String{Value: string(bytes)}
	}

	for _, dns := range talosKubelet.KubeletClusterDNS {
		kubelet.ClusterDNS = append(kubelet.ClusterDNS, types.String{Value: dns})
	}

	for _, mount := range talosKubelet.KubeletExtraMounts {
		extraMount := ExtraMount{}
		err := extraMount.Read(mount)
		if err != nil {
			return err
		}
		kubelet.ExtraMounts = append(kubelet.ExtraMounts, extraMount)
	}

	if !reflect.DeepEqual(talosKubelet.KubeletNodeIP, v1alpha1.KubeletNodeIPConfig{}) {
		for _, subnet := range talosKubelet.KubeletNodeIP.KubeletNodeIPValidSubnets {
			kubelet.NodeIPValidSubnets = append(kubelet.NodeIPValidSubnets, types.String{Value: subnet})
		}
	}

	return nil
}

type TalosKubelet struct {
	*v1alpha1.KubeletConfig
}

func (talosKubelet TalosKubelet) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) error {
			if talosKubelet.KubeletConfig == nil {
				return nil
			}

			if planConfig.Kubelet == nil {
				planConfig.Kubelet = &KubeletConfig{}
			}

			if talosKubelet.KubeletImage != (&v1alpha1.KubeletConfig{}).Image() {
				planConfig.Kubelet.Image = readString(talosKubelet.Image())
			}
			if planConfig.Kubelet.Image.Value == "" {
				planConfig.Kubelet.Image.Value = (&v1alpha1.KubeletConfig{}).Image()
			}

			planConfig.Kubelet.RegisterWithFQDN = readBool(talosKubelet.RegisterWithFQDN())

			var obj *types.String
			var err error
			if obj, err = readObject(talosKubelet.KubeletExtraConfig); err != nil {
				return err
			}
			if obj != nil {
				planConfig.Kubelet.ExtraConfig = *obj
			}

			planConfig.Kubelet.ClusterDNS = readStringList(talosKubelet.ClusterDNS())
			planConfig.Kubelet.ExtraArgs = readStringMap(talosKubelet.ExtraArgs())
			planConfig.Kubelet.NodeIPValidSubnets = readStringList(talosKubelet.NodeIP().ValidSubnets())

			if planConfig.Kubelet.zero() {
				planConfig.Kubelet = nil
			}

			return nil
		},
	}

	funs = append(funs, TalosExtraMounts{ExtraMounts: &talosKubelet.KubeletExtraMounts}.ReadFunc()...)

	return funs
}

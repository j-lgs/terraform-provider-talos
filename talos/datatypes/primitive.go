package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

type ExtraManifestHeaders map[string]types.String

func (headers ExtraManifestHeaders) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringMap(headers, &cfg.ClusterConfig.ExtraManifestHeaders)
			return nil
		},
	}
}

type TalosExtraManifestHeaders map[string]string

func (talosExtraManifestHeaders TalosExtraManifestHeaders) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if len(talosExtraManifestHeaders) <= 0 {
				return nil
			}

			planConfig.ExtraManifestHeaders = readStringMap(talosExtraManifestHeaders)
			return nil
		},
	}
	return funs
}

type MachineSysfs map[string]types.String

func (sysfs MachineSysfs) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringMap(sysfs, &cfg.MachineConfig.MachineSysfs)
			return nil
		},
	}
}

type TalosMachineSysfs map[string]string

func (talosMachineSysfs TalosMachineSysfs) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if len(talosMachineSysfs) <= 0 {
				return nil
			}

			planConfig.Sysfs = readStringMap(talosMachineSysfs)
			return nil
		},
	}
	return funs
}

type MachineSysctls map[string]types.String

func (sysctls MachineSysctls) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringMap(sysctls, &cfg.MachineConfig.MachineSysctls)
			return nil
		},
	}
}

type TalosMachineSysctls map[string]string

func (talosMachineSysctls TalosMachineSysctls) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if len(talosMachineSysctls) <= 0 {
				return nil
			}

			planConfig.Sysctls = readStringMap(talosMachineSysctls)
			return nil
		},
	}
	return funs
}

type MachineEnv map[string]types.String

func (env MachineEnv) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringMap(env, &cfg.MachineConfig.MachineEnv)
			return nil
		},
	}
}

type TalosMachineEnv map[string]string

func (talosMachineEnv TalosMachineEnv) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if len(talosMachineEnv) <= 0 {
				return nil
			}

			planConfig.Env = readStringMap(talosMachineEnv)
			return nil
		},
	}
	return funs
}

type MachineCertSANs []types.String

func (certSANs MachineCertSANs) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringList(certSANs, &cfg.MachineConfig.MachineCertSANs)
			return nil
		},
	}
}

type TalosMachineCertSANs []string

func (talosMachineCertSANs TalosMachineCertSANs) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if len(talosMachineCertSANs) <= 0 {
				return nil
			}

			planConfig.CertSANS = readStringList(talosMachineCertSANs)
			return nil
		},
	}
	return funs
}

type MachineUdevRules []types.String

func (rules MachineUdevRules) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			if cfg.MachineConfig.MachineUdev == nil {
				cfg.MachineConfig.MachineUdev = &v1alpha1.UdevConfig{}
			}

			setStringList(rules, &cfg.MachineConfig.MachineUdev.UdevRules)
			return nil
		},
	}
}

type TalosMachineUdev []string

func (talosMachineUdev TalosMachineUdev) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			planConfig.Udev = readStringList(talosMachineUdev)
			return nil
		},
	}
	return funs
}

type MachinePods []types.String

func (pods MachinePods) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setObjectList(pods, &cfg.MachineConfig.MachinePods)
			return nil
		},
	}
}

type TalosMachinePods []v1alpha1.Unstructured

func (talosMachinePods TalosMachinePods) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if len(talosMachinePods) <= 0 {
				return nil
			}

			planConfig.Pod, err = readObjectList(talosMachinePods)
			if err != nil {
				return
			}

			return nil
		},
	}
	return funs
}

type ClusterExtraManifests []types.String

func (manifests ClusterExtraManifests) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringList(manifests, &cfg.ClusterConfig.ExtraManifests)
			return nil
		},
	}
}

type TalosClusterExtraManifests []string

func (talosMachineExtraManfests TalosClusterExtraManifests) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if len(talosMachineExtraManfests) <= 0 {
				return nil
			}

			planConfig.ExtraManifests = readStringList(talosMachineExtraManfests)
			return nil
		},
	}
	return funs
}

type ExternalCloudProvider []types.String

func (manifests ExternalCloudProvider) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			if cfg.ClusterConfig.ExternalCloudProviderConfig == nil {
				cfg.ClusterConfig.ExternalCloudProviderConfig =
					&v1alpha1.ExternalCloudProviderConfig{}
			}

			if len(manifests) > 0 {
				cfg.ClusterConfig.ExternalCloudProviderConfig.ExternalEnabled = true
			}

			setStringList(manifests, &cfg.ClusterConfig.ExternalCloudProviderConfig.ExternalManifests)
			return nil
		},
	}
}

type TalosExternalCloudProvider []string

func (talosExternalCloudProvider TalosExternalCloudProvider) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			planConfig.ExternalCloudProvider = readStringList(talosExternalCloudProvider)
			return nil
		},
	}
	return funs
}

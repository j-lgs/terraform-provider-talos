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

type MachineSysfs map[string]types.String

func (sysfs MachineSysfs) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringMap(sysfs, &cfg.MachineConfig.MachineSysfs)
			return nil
		},
	}
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

type MachineEnv map[string]types.String

func (env MachineEnv) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setStringMap(env, &cfg.MachineConfig.MachineEnv)
			return nil
		},
	}
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

type MachinePods []types.String

func (pods MachinePods) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			setObjectList(pods, &cfg.MachineConfig.MachinePods)
			return nil
		},
	}
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

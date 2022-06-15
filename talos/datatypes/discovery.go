package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planDiscovery ClusterDiscoveryConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	funcs := [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			discovery := cfg.ClusterConfig.ClusterDiscoveryConfig

			setBool(planDiscovery.Enabled, &discovery.DiscoveryEnabled)

			return nil
		},
	}

	if planDiscovery.Registries != nil {
		funcs = append(funcs, planDiscovery.Registries.DataFunc()...)
	}

	return funcs
}

type TalosClusterDiscoveryConfig struct {
	*v1alpha1.ClusterDiscoveryConfig
}

func (talosClusterDiscoveryConfig TalosClusterDiscoveryConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if talosClusterDiscoveryConfig.ClusterDiscoveryConfig == nil {
				return nil
			}

			if !talosClusterDiscoveryConfig.DiscoveryEnabled {
				return nil
			}

			if planConfig.Discovery == nil {
				planConfig.Discovery = &ClusterDiscoveryConfig{}
			}

			planConfig.Discovery.Enabled = readBool(talosClusterDiscoveryConfig.DiscoveryEnabled)

			return nil
		},
	}

	funs = append(funs, TalosDiscoveryRegistriesConfig{DiscoveryRegistriesConfig: talosClusterDiscoveryConfig.DiscoveryRegistries}.ReadFunc()...)

	return funs
}

func (planDiscoveryRegistries DiscoveryRegistriesConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			registry := &cfg.ClusterConfig.ClusterDiscoveryConfig.DiscoveryRegistries

			setBool(planDiscoveryRegistries.KubernetesDisabled, &registry.RegistryKubernetes.RegistryDisabled)
			setBool(planDiscoveryRegistries.ServiceDisabled, &registry.RegistryService.RegistryDisabled)

			setString(planDiscoveryRegistries.ServiceEndpoint, &registry.RegistryService.RegistryEndpoint)

			return nil
		},
	}
}

type TalosDiscoveryRegistriesConfig struct {
	v1alpha1.DiscoveryRegistriesConfig
}

func (talosDiscoveryRegistriesConfig TalosDiscoveryRegistriesConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.Discovery.Registries == nil {
				planConfig.Discovery.Registries = &DiscoveryRegistriesConfig{}
			}

			planConfig.Discovery.Registries.KubernetesDisabled = readBool(talosDiscoveryRegistriesConfig.RegistryKubernetes.RegistryDisabled)
			planConfig.Discovery.Registries.ServiceDisabled = readBool(talosDiscoveryRegistriesConfig.RegistryService.RegistryDisabled)
			planConfig.Discovery.Registries.ServiceEndpoint = readString(talosDiscoveryRegistriesConfig.RegistryService.RegistryEndpoint)

			return nil
		},
	}
	return funs
}

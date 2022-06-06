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

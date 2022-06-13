package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planKernelConfig KernelConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			if len(planKernelConfig.Modules) <= 0 {
				return nil
			}

			modules := []*v1alpha1.KernelModuleConfig{}
			for _, module := range planKernelConfig.Modules {
				modules = append(modules, &v1alpha1.KernelModuleConfig{ModuleName: module.Value})
			}

			kernelConfig := &v1alpha1.KernelConfig{KernelModules: modules}

			cfg.MachineConfig.MachineKernel = kernelConfig
			return nil
		},
	}
}

type KernelModules = []*v1alpha1.KernelModuleConfig
type TalosKernelModuleConfigs struct {
	*KernelModules
}

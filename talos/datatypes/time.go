package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planTimeConfig TimeConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			timeConfig := &v1alpha1.TimeConfig{
				TimeDisabled: planTimeConfig.Disabled.Value,
			}

			setStringList(planTimeConfig.Servers, &timeConfig.TimeServers)

			if err := setStringDuration(planTimeConfig.BootTimeout, &timeConfig.TimeBootTimeout); err != nil {
				return err
			}

			cfg.MachineConfig.MachineTime = timeConfig
			return nil
		},
	}
}

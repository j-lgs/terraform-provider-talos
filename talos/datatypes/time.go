package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planTimeConfig TimeConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			timeConfig := cfg.MachineConfig.MachineTime

			if timeConfig == nil {
				timeConfig = &v1alpha1.TimeConfig{}
			}

			setBool(planTimeConfig.Disabled, &timeConfig.TimeDisabled)
			setStringList(planTimeConfig.Servers, &timeConfig.TimeServers)

			if err := setStringDuration(planTimeConfig.BootTimeout, &timeConfig.TimeBootTimeout); err != nil {
				return err
			}

			cfg.MachineConfig.MachineTime = timeConfig
			return nil
		},
	}
}

type TalosTimeConfig struct {
	*v1alpha1.TimeConfig
}

func (talosTimeConfig TalosTimeConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if talosTimeConfig.TimeConfig == nil {
				return nil
			}

			if planConfig.Time == nil {
				planConfig.Time = &TimeConfig{}
			}

			planConfig.Time.BootTimeout = readStringDuration(talosTimeConfig.TimeBootTimeout)
			planConfig.Time.Disabled = readBool(talosTimeConfig.TimeDisabled)
			planConfig.Time.Servers = readStringList(talosTimeConfig.TimeServers)

			return nil
		},
	}
	return funs
}

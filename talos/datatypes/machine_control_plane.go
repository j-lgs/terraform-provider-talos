package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planMachineControlPlane MachineControlPlane) Data() (interface{}, error) {
	controlConf := &v1alpha1.MachineControlPlaneConfig{
		MachineControllerManager: &v1alpha1.MachineControllerManagerConfig{},
		MachineScheduler:         &v1alpha1.MachineSchedulerConfig{},
	}

	if !planMachineControlPlane.ControllerManagerDisabled.Null {
		controlConf.MachineControllerManager.MachineControllerManagerDisabled = planMachineControlPlane.ControllerManagerDisabled.Value
	}

	if !planMachineControlPlane.ControllerManagerDisabled.Null {
		controlConf.MachineScheduler.MachineSchedulerDisabled = planMachineControlPlane.SchedulerDisabled.Value
	}

	return controlConf, nil
}

func (machineControlPlane MachineControlPlane) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			mcp, err := machineControlPlane.Data()
			if err != nil {
				return err
			}
			cfg.MachineConfig.MachineControlPlane = mcp.(*v1alpha1.MachineControlPlaneConfig)
			return nil
		},
	}
}

type TalosMCPConfig struct {
	*v1alpha1.MachineControlPlaneConfig
}

func (talosMCPConfig TalosMCPConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.MachineControlPlane == nil {
				planConfig.MachineControlPlane = &MachineControlPlane{}
			}

			return nil
		},
	}
	return funs
}

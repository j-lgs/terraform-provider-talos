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

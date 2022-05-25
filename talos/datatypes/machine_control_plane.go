package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// MachineControlPlaneSchema configures options pertaining to the Kubernetes control plane that's installed onto the machine.
var MachineControlPlaneSchema tfsdk.Schema = tfsdk.Schema{
	Description: "Configures options pertaining to the Kubernetes control plane that's installed onto the machine",
	Attributes: map[string]tfsdk.Attribute{
		"controller_manager_disabled": {
			Type:     types.BoolType,
			Optional: true,
			Description: "Disable kube-controller-manager on the node.	",
		},
		"scheduler_disabled": {
			Type:        types.BoolType,
			Optional:    true,
			Description: "Disable kube-scheduler on the node.",
		},
	},
}

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

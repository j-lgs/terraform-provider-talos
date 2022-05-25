package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planFile File) Data() (interface{}, error) {
	return &v1alpha1.MachineFile{
		FileContent:     planFile.Content.Value,
		FilePermissions: v1alpha1.FileMode(planFile.Permissions.Value),
		FilePath:        planFile.Path.Value,
		FileOp:          planFile.Op.Value,
	}, nil
}

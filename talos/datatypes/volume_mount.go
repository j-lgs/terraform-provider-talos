package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (mount VolumeMount) Data() (interface{}, error) {
	vol := v1alpha1.VolumeMountConfig{
		VolumeHostPath:  mount.HostPath.Value,
		VolumeMountPath: mount.MountPath.Value,
	}
	if !mount.Readonly.Null {
		vol.VolumeReadOnly = mount.Readonly.Value
	}

	return vol, nil
}

// Read copies data from talos types to terraform state types.
func (mount *VolumeMount) Read(vol interface{}) error {
	volume := vol.(v1alpha1.VolumeMountConfig)

	mount.HostPath = types.String{Value: volume.VolumeHostPath}
	mount.MountPath = types.String{Value: volume.VolumeMountPath}
	mount.Readonly = types.Bool{Value: volume.VolumeReadOnly}

	return nil
}

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

type VolumeMounts []v1alpha1.VolumeMountConfig
type TalosKubeletMounts struct {
	VolumeMounts
}

func (talosVolumeMounts TalosKubeletMounts) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.APIServer.ExtraVolumes == nil {
				planConfig.APIServer.ExtraVolumes = make([]VolumeMount, 0)
			}

			for _, mount := range talosVolumeMounts.VolumeMounts {
				planConfig.APIServer.ExtraVolumes = append(planConfig.APIServer.ExtraVolumes, VolumeMount{
					HostPath:  readString(mount.HostPath()),
					MountPath: readString(mount.MountPath()),
					Readonly:  readBool(mount.ReadOnly()),
				})
			}

			return nil
		},
	}
	return funs
}

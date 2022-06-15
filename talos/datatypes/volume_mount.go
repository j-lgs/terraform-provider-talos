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
type TalosAPIServerMounts struct {
	VolumeMounts
}

func readExtraVolumes(mounts []VolumeMount, talosMounts VolumeMounts) []VolumeMount {
	if mounts == nil {
		mounts = make([]VolumeMount, 0)
	}

	for _, mount := range talosMounts {
		mounts = append(mounts, VolumeMount{
			HostPath:  readString(mount.HostPath()),
			MountPath: readString(mount.MountPath()),
			Readonly:  readBool(mount.ReadOnly()),
		})
	}

	return mounts
}

func (talosVolumeMounts TalosAPIServerMounts) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) error {
			planConfig.APIServer.ExtraVolumes = readExtraVolumes(planConfig.APIServer.ExtraVolumes,
				talosVolumeMounts.VolumeMounts)
			return nil
		},
	}
	return funs
}

type TalosSchedulerMounts struct {
	VolumeMounts
}

func (talosVolumeMounts TalosSchedulerMounts) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) error {
			planConfig.Scheduler.ExtraVolumes = readExtraVolumes(planConfig.Scheduler.ExtraVolumes,
				talosVolumeMounts.VolumeMounts)
			return nil
		},
	}
	return funs
}

type TalosControllerManagerMounts struct {
	VolumeMounts
}

func (talosVolumeMounts TalosControllerManagerMounts) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) error {
			planConfig.ControllerManager.ExtraVolumes = readExtraVolumes(planConfig.ControllerManager.ExtraVolumes,
				talosVolumeMounts.VolumeMounts)
			return nil
		},
	}
	return funs
}

package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planControllerManager ControllerManagerConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			controllerManager := cfg.ClusterConfig.ControllerManagerConfig

			if planControllerManager.Image.Null {
				controllerManager.ContainerImage = (&v1alpha1.ControllerManagerConfig{}).Image()
			}
			setString(planControllerManager.Image, &cfg.ClusterConfig.ControllerManagerConfig.ContainerImage)

			setStringMap(planControllerManager.ExtraArgs, &controllerManager.ExtraArgsConfig)
			setStringMap(planControllerManager.Env, &controllerManager.EnvConfig)

			// TODO: Migrate to datafuncs
			if len(planControllerManager.ExtraVolumes) > 0 &&
				len(controllerManager.ExtraVolumesConfig) == 0 {
				controllerManager.ExtraVolumesConfig = []v1alpha1.VolumeMountConfig{}
			}
			for _, mount := range planControllerManager.ExtraVolumes {
				m, err := mount.Data()
				if err != nil {
					return err
				}
				controllerManager.ExtraVolumesConfig = append(controllerManager.ExtraVolumesConfig, m.(v1alpha1.VolumeMountConfig))
			}

			return nil
		},
	}
}

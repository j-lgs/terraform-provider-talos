package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planScheduler SchedulerConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			scheduler := cfg.ClusterConfig.SchedulerConfig

			if planScheduler.Image.Null {
				scheduler.ContainerImage = (&v1alpha1.SchedulerConfig{}).Image()
			}

			setStringMap(planScheduler.ExtraArgs, &scheduler.ExtraArgsConfig)
			setStringMap(planScheduler.Env, &scheduler.EnvConfig)

			// TODO: Migrate to datafuncs
			if err := setVolumeMounts(planScheduler.ExtraVolumes, &scheduler.ExtraVolumesConfig); err != nil {
				return err
			}

			return nil
		},
	}
}

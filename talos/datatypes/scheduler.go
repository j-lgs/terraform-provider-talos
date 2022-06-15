package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planScheduler SchedulerConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			scheduler := cfg.ClusterConfig.SchedulerConfig

			if scheduler == nil {
				scheduler = &v1alpha1.SchedulerConfig{}
			}

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

func (planScheduler SchedulerConfig) zero() bool {
	return mkString(planScheduler.Image).zero() &&
		len(planScheduler.Env) <= 0 &&
		len(planScheduler.ExtraArgs) <= 0 &&
		len(planScheduler.ExtraVolumes) <= 0
}

type TalosSchedulerConfig struct {
	*v1alpha1.SchedulerConfig
}

func (talosSchedulerConfig TalosSchedulerConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.Scheduler == nil {
				planConfig.Scheduler = &SchedulerConfig{}
			}

			if planConfig.Scheduler.zero() {
				planConfig.Scheduler = nil
				return nil
			}
			return nil
		},
	}
	return funs
}

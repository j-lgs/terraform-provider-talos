package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planAdminKubeconfig AdminKubeconfigConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			if cfg.ClusterConfig.AdminKubeconfigConfig == nil {
				cfg.ClusterConfig.AdminKubeconfigConfig =
					&v1alpha1.AdminKubeconfigConfig{}
			}

			if err := setStringDuration(planAdminKubeconfig.CertLifetime,
				&cfg.ClusterConfig.AdminKubeconfigConfig.AdminKubeconfigCertLifetime); err != nil {
				return err
			}

			return nil
		},
	}
}

type TalosAdminKubeconfigConfig struct {
	*v1alpha1.AdminKubeconfigConfig
}

func (talosKubeconfig TalosAdminKubeconfigConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.AdminKubeConfig == nil {
				planConfig.AdminKubeConfig = &AdminKubeconfigConfig{}
			}

			planConfig.AdminKubeConfig.CertLifetime = readStringDuration(talosKubeconfig.CertLifetime())

			return nil
		},
	}
	return funs
}

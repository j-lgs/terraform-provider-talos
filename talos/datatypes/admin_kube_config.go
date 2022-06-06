package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planAdminKubeconfig AdminKubeconfigConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			adminKubeconfig := cfg.ClusterConfig.AdminKubeconfigConfig
			if adminKubeconfig == nil {
				adminKubeconfig = &v1alpha1.AdminKubeconfigConfig{}
			}

			setStringDuration(planAdminKubeconfig.CertLifetime, &adminKubeconfig.AdminKubeconfigCertLifetime)

			return nil
		},
	}
}

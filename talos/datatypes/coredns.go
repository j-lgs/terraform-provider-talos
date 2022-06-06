package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planCoreDNS CoreDNS) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			if cfg.ClusterConfig.CoreDNSConfig == nil {
				cfg.ClusterConfig.CoreDNSConfig = &v1alpha1.CoreDNS{}
			}

			if planCoreDNS.Image.Null {
				cfg.ClusterConfig.CoreDNSConfig.CoreDNSImage =
					(&v1alpha1.CoreDNS{}).Image()
			}
			setString(planCoreDNS.Image, &cfg.ClusterConfig.CoreDNSConfig.CoreDNSImage)

			setBool(planCoreDNS.Disabled,
				&cfg.ClusterConfig.CoreDNSConfig.CoreDNSDisabled)

			return nil
		},
	}
}

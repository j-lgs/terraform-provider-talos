package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planCoreDNS CoreDNS) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			coreDNS := cfg.ClusterConfig.CoreDNSConfig

			if coreDNS == nil {
				coreDNS = &v1alpha1.CoreDNS{}
			}

			if planCoreDNS.Image.Null {
				coreDNS.CoreDNSImage = (&v1alpha1.CoreDNS{}).Image()
			}

			setBool(planCoreDNS.Disabled, &coreDNS.CoreDNSDisabled)

			return nil
		},
	}
}

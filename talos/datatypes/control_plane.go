package datatypes

import (
	"net/url"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planControlPlane ControlPlaneConfig) Data() (any, error) {
	controlplane := &v1alpha1.ControlPlaneConfig{}

	url, err := url.Parse(planControlPlane.Endpoint.Value)
	if err != nil {
		return &v1alpha1.ControlPlaneConfig{}, err
	}

	controlplane.Endpoint = &v1alpha1.Endpoint{
		URL: url,
	}

	if !planControlPlane.LocalAPIServerPort.Null {
		controlplane.LocalAPIServerPort = int(planControlPlane.LocalAPIServerPort.Value)
	}

	return controlplane, nil
}

func (planControlPlane ControlPlaneConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			ins, err := planControlPlane.Data()
			if err != nil {
				return err
			}
			cfg.ClusterConfig.ControlPlane = ins.(*v1alpha1.ControlPlaneConfig)
			return nil
		},
	}
}

package datatypes

import (
	"net/url"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planLoggingConfig LoggingConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	ptdfs := []PlanToDataFunc{}
	for _, v := range planLoggingConfig.Destinations {
		ptdfs = append(ptdfs, v)
	}

	return AppendDataFunc([]ConfigDataFunc{}, ptdfs...)
}

func (planLoggingDestination LoggingDestination) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			if cfg.MachineConfig.MachineLogging == nil {
				cfg.MachineConfig.MachineLogging = &v1alpha1.LoggingConfig{}
			}

			endpoint, err := url.Parse(planLoggingDestination.Endpoint.Value)
			if err != nil {
				return err
			}

			cfg.MachineConfig.MachineLogging.LoggingDestinations =
				append(cfg.MachineConfig.MachineLogging.LoggingDestinations,
					v1alpha1.LoggingDestination{
						LoggingEndpoint: &v1alpha1.Endpoint{URL: endpoint},
						LoggingFormat:   planLoggingDestination.Format.Value,
					})

			return nil
		},
	}
}

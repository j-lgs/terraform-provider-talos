package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planRoute Route) Data() (interface{}, error) {
	route := &v1alpha1.Route{
		RouteNetwork: planRoute.Network.Value,
	}

	if !planRoute.Gateway.Null {
		route.RouteGateway = planRoute.Gateway.Value
	}
	if !planRoute.Source.Null {
		route.RouteSource = planRoute.Source.Value
	}
	if !planRoute.Metric.Null {
		route.RouteMetric = uint32(planRoute.Metric.Value)
	}

	return route, nil
}

package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
// TODO add DHCPUIDv6 from struct, and add it to the Talos config documentation.
func (planDHCPOptions DHCPOptions) Data() (interface{}, error) {
	dhcpOptions := &v1alpha1.DHCPOptions{}

	if !planDHCPOptions.IPV4.Null {
		dhcpOptions.DHCPIPv4 = &planDHCPOptions.IPV4.Value
	}
	if !planDHCPOptions.IPV6.Null {
		dhcpOptions.DHCPIPv6 = &planDHCPOptions.IPV6.Value
	}
	if !planDHCPOptions.RouteMetric.Null {
		dhcpOptions.DHCPRouteMetric = uint32(planDHCPOptions.RouteMetric.Value)
	}

	return dhcpOptions, nil
}

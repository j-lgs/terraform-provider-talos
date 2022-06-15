package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config"
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

func readDHCP(dhcp config.DHCPOptions) (out *DHCPOptions) {
	out = &DHCPOptions{}

	out.IPV4 = readBool(dhcp.IPv4())
	out.IPV6 = readBool(dhcp.IPv6())
	out.RouteMetric = readInt(int(dhcp.RouteMetric()))

	return
}

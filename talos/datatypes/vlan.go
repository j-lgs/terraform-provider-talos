package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planVLAN VLAN) Data() (interface{}, error) {
	vlan := &v1alpha1.Vlan{}

	for _, vlanAddress := range planVLAN.Addresses {
		vlan.VlanAddresses = append(vlan.VlanAddresses, vlanAddress.Value)
	}
	for _, planVLANRoute := range planVLAN.Routes {
		route, err := planVLANRoute.Data()
		if err != nil {
			return &v1alpha1.Vlan{}, err
		}
		vlan.VlanRoutes = append(vlan.VlanRoutes, route.(*v1alpha1.Route))
	}
	if !planVLAN.DHCP.Null {
		vlan.VlanDHCP = planVLAN.DHCP.Value
	}
	if !planVLAN.VLANId.Null {
		vlan.VlanID = uint16(planVLAN.VLANId.Value)
	}
	if !planVLAN.MTU.Null {
		vlan.VlanMTU = uint32(planVLAN.MTU.Value)
	}
	if planVLAN.VIP != nil {
		vip, err := planVLAN.VIP.Data()
		if err != nil {
			return &v1alpha1.Vlan{}, err
		}
		vlan.VlanVIP = vip.(*v1alpha1.DeviceVIPConfig)
	}
	return vlan, nil
}

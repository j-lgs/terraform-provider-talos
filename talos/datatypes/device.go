package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planDevice NetworkDevice) Data() (interface{}, error) {
	device := &v1alpha1.Device{}

	device.DeviceInterface = planDevice.Name.Value

	for _, address := range planDevice.Addresses {
		device.DeviceAddresses = append(device.DeviceAddresses, address.Value)
	}

	for _, planRoute := range planDevice.Routes {
		route, err := planRoute.Data()
		if err != nil {
			return &v1alpha1.Device{}, err
		}
		device.DeviceRoutes = append(device.DeviceRoutes, route.(*v1alpha1.Route))
	}

	if planDevice.BondData != nil {
		bond, err := planDevice.BondData.Data()
		if err != nil {
			return nil, err
		}
		device.DeviceBond = bond.(*v1alpha1.Bond)
	}

	if !planDevice.DHCP.Null {
		device.DeviceDHCP = planDevice.DHCP.Value
	}

	if planDevice.DHCPOptions != nil {
		dhcpopts, err := planDevice.DHCPOptions.Data()
		if err != nil {
			return &v1alpha1.Device{}, err
		}
		device.DeviceDHCPOptions = dhcpopts.(*v1alpha1.DHCPOptions)
	}

	for _, planVLAN := range planDevice.VLANs {
		vlan, err := planVLAN.Data()
		if err != nil {
			return &v1alpha1.Device{}, err
		}
		device.DeviceVlans = append(device.DeviceVlans, vlan.(*v1alpha1.Vlan))
	}

	if !planDevice.MTU.Null {
		device.DeviceMTU = int(planDevice.MTU.Value)
	}

	if !planDevice.DHCP.Null {
		device.DeviceDHCP = planDevice.DHCP.Value
	}

	if !planDevice.Ignore.Null {
		device.DeviceIgnore = planDevice.Ignore.Value
	}

	if !planDevice.Dummy.Null {
		device.DeviceDummy = planDevice.Dummy.Value
	}
	if planDevice.Wireguard != nil {
		wireguard, err := planDevice.Wireguard.Data()
		if err != nil {
			return v1alpha1.Device{}, err
		}
		device.DeviceWireguardConfig = wireguard.(*v1alpha1.DeviceWireguardConfig)
	}
	if planDevice.VIP != nil {
		vip, err := planDevice.VIP.Data()
		if err != nil {
			return v1alpha1.Device{}, err
		}
		device.DeviceVIPConfig = vip.(*v1alpha1.DeviceVIPConfig)
	}

	return device, nil
}

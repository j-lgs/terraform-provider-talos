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

type TalosNetworkInterface struct {
	*v1alpha1.Device
}

func readInterface(talosNetworkInterface TalosNetworkInterface, device *NetworkDevice) (err error) {
	device.Name = readString(talosNetworkInterface.Interface())
	device.Addresses = readStringList(talosNetworkInterface.Addresses())

	device.DHCP = readBool(talosNetworkInterface.DHCP())
	device.Dummy = readBool(talosNetworkInterface.Dummy())
	device.Ignore = readBool(talosNetworkInterface.Ignore())
	device.MTU = readInt(talosNetworkInterface.MTU())

	for _, route := range talosNetworkInterface.Routes() {
		device.Routes = append(device.Routes, readRoute(route))
	}

	if talosNetworkInterface.DeviceBond != nil {
		device.BondData = readBond(talosNetworkInterface.Bond())
	}

	if talosNetworkInterface.DeviceDHCPOptions != nil {
		device.DHCPOptions = readDHCP(talosNetworkInterface.DHCPOptions())
	}

	if talosNetworkInterface.DeviceVIPConfig != nil {
		device.VIP = readVIPConfig(talosNetworkInterface.VIPConfig())
	}

	for _, vlan := range talosNetworkInterface.Vlans() {
		device.VLANs = append(device.VLANs, readVLAN(vlan))
	}

	if talosNetworkInterface.DeviceWireguardConfig != nil {
		device.Wireguard, err = readWireguardConfig(talosNetworkInterface.WireguardConfig())
		if err != nil {
			return
		}
	}

	return
}

func (talosNetworkInterface TalosNetworkInterface) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			inList := false
			for _, device := range planConfig.Network.Devices {
				if device.Name.Value == talosNetworkInterface.Interface() {
					readInterface(talosNetworkInterface, &device)
					inList = true
				}
			}

			if !inList {
				device := NetworkDevice{}
				readInterface(talosNetworkInterface, &device)
				planConfig.Network.Devices = append(planConfig.Network.Devices, device)
			}

			return nil
		},
	}
	return funs
}

type Interfaces []*v1alpha1.Device

type TalosNetworkInterfaces struct {
	Interfaces
}

func (talosNetworkInterfaces TalosNetworkInterfaces) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{}

	for _, device := range talosNetworkInterfaces.Interfaces {
		funs = append(funs, TalosNetworkInterface{Device: device}.ReadFunc()...)
	}

	return funs
}

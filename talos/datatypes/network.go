package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planNetwork NetworkConfig) Data() (any, error) {
	network := &v1alpha1.NetworkConfig{
		NetworkHostname: planNetwork.Hostname.Value,
	}

	for _, device := range planNetwork.Devices {
		dev, err := device.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		network.NetworkInterfaces = append(network.NetworkInterfaces, dev.(*v1alpha1.Device))
	}

	for _, ns := range planNetwork.Nameservers {
		network.NameServers = append(network.NameServers, ns.Value)
	}

	network.ExtraHostEntries = []*v1alpha1.ExtraHost{}
	for hostname, addresses := range planNetwork.ExtraHosts {
		host := &v1alpha1.ExtraHost{
			HostIP: hostname,
		}
		network.ExtraHostEntries = append(network.ExtraHostEntries, host)
		for _, address := range addresses {
			host.HostAliases = append(host.HostAliases, address.Value)
		}
	}

	if planNetwork.Kubespan != nil {
		kubespan, err := planNetwork.Kubespan.Data()
		if err != nil {
			return &v1alpha1.Config{}, err
		}
		network.NetworkKubeSpan = kubespan.(v1alpha1.NetworkKubeSpan)
	}

	return network, nil
}

func (planNetwork NetworkConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			ins, err := planNetwork.Data()
			if err != nil {
				return err
			}
			cfg.MachineConfig.MachineNetwork = ins.(*v1alpha1.NetworkConfig)
			return nil
		},
	}
}

type TalosNetworkConfig struct {
	*v1alpha1.NetworkConfig
}

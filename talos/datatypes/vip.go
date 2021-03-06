package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planVIP VIP) Data() (interface{}, error) {
	vip := &v1alpha1.DeviceVIPConfig{
		SharedIP: planVIP.IP.Value,
	}
	if !planVIP.EquinixMetalAPIToken.Null {
		vip.EquinixMetalConfig = &v1alpha1.VIPEquinixMetalConfig{
			EquinixMetalAPIToken: planVIP.EquinixMetalAPIToken.Value,
		}
	}
	if !planVIP.HetznerCloudAPIToken.Null {
		vip.HCloudConfig = &v1alpha1.VIPHCloudConfig{
			HCloudAPIToken: planVIP.HetznerCloudAPIToken.Value,
		}
	}

	return vip, nil
}

func readVIPConfig(vip config.VIPConfig) (out *VIP) {
	out = &VIP{}

	out.EquinixMetalAPIToken = readString(vip.EquinixMetal().APIToken())
	out.HetznerCloudAPIToken = readString(vip.HCloud().APIToken())
	out.IP = readString(vip.IP())

	return
}

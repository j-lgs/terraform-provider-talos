package datatypes

import (
	"time"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planWireguard Wireguard) Data() (interface{}, error) {
	wireguard := &v1alpha1.DeviceWireguardConfig{}

	for _, planPeer := range planWireguard.Peers {
		peer, err := planPeer.Data()
		if err != nil {
			return &v1alpha1.DeviceWireguardConfig{}, nil
		}
		wireguard.WireguardPeers = append(wireguard.WireguardPeers, peer.(*v1alpha1.DeviceWireguardPeer))
	}

	if !planWireguard.PrivateKey.Null {
		wireguard.WireguardPrivateKey = planWireguard.PrivateKey.Value
	}

	if !planWireguard.FirewallMark.Null {
		wireguard.WireguardFirewallMark = int(planWireguard.FirewallMark.Value)
	}

	if !planWireguard.ListenPort.Null {
		wireguard.WireguardListenPort = int(planWireguard.ListenPort.Value)
	}

	return wireguard, nil
}

// Data copies data from terraform state types to talos types.
func (planPeer WireguardPeer) Data() (interface{}, error) {
	peer := &v1alpha1.DeviceWireguardPeer{
		WireguardPublicKey: planPeer.PublicKey.Value,
		WireguardEndpoint:  planPeer.Endpoint.Value,
	}

	for _, ip := range planPeer.AllowedIPs {
		peer.WireguardAllowedIPs = append(peer.WireguardAllowedIPs, ip.Value)
	}

	if !planPeer.PersistentKeepaliveInterval.Null {
		peer.WireguardPersistentKeepaliveInterval = time.Duration(planPeer.PersistentKeepaliveInterval.Value) * time.Second
	}

	return peer, nil
}

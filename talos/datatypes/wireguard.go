package datatypes

import (
	"time"

	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

func readWireguardPeer(peer config.WireguardPeer) (out WireguardPeer) {
	out.AllowedIPs = readStringList(peer.AllowedIPs())
	out.Endpoint = readString(peer.Endpoint())
	out.PersistentKeepaliveInterval = readInt(int(peer.PersistentKeepaliveInterval().Seconds()))
	out.PublicKey = readString(peer.PublicKey())

	return
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

func readWireguardConfig(config config.WireguardConfig) (out *Wireguard, err error) {
	out = &Wireguard{}

	out.FirewallMark = readInt(config.FirewallMark())
	out.ListenPort = readInt(config.ListenPort())
	out.PrivateKey = readString(config.PrivateKey())

	key, err := wgtypes.ParseKey(config.PrivateKey())
	if err != nil {
		return
	}
	out.PublicKey = readString(key.PublicKey().String())

	for _, peer := range config.Peers() {
		out.Peers = append(out.Peers, readWireguardPeer(peer))
	}

	return
}

package datatypes

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planKubespan NetworkKubeSpan) Data() (any, error) {
	kubespan := v1alpha1.NetworkKubeSpan{}

	mkBool(planKubespan.Enabled).set(&kubespan.KubeSpanEnabled)
	mkBool(planKubespan.AllowPeerDownBypass).set(&kubespan.KubeSpanAllowDownPeerBypass)

	return kubespan, nil
}

func (planKubespan NetworkKubeSpan) zero() bool {
	return mkBool(planKubespan.Enabled).zero() && mkBool(planKubespan.AllowPeerDownBypass).zero()
}

func (stateKubespan *NetworkKubeSpan) Read(kubespan any) error {
	if kubespan == nil {
		return fmt.Errorf("nil talos NetworkKubeSpan pointer provided to Read function")
	}
	networkKubeSpan := kubespan.(*v1alpha1.NetworkKubeSpan)

	stateKubespan.Enabled.Value = networkKubeSpan.KubeSpanEnabled

	if networkKubeSpan.KubeSpanAllowDownPeerBypass {
		stateKubespan.AllowPeerDownBypass = types.Bool{Value: networkKubeSpan.KubeSpanAllowDownPeerBypass}
	}

	return nil
}

type TalosNetworkKubeSpan struct {
	v1alpha1.NetworkKubeSpan
}

func (talosKubeSpan TalosNetworkKubeSpan) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.Network.Kubespan == nil {
				planConfig.Network.Kubespan = &NetworkKubeSpan{}
			}

			mkBool(talosKubeSpan.Enabled()).read(&planConfig.Network.Kubespan.Enabled)
			mkBool(talosKubeSpan.KubeSpanAllowDownPeerBypass).read(&planConfig.Network.Kubespan.AllowPeerDownBypass)

			if planConfig.Network.Kubespan.zero() {
				planConfig.Network.Kubespan = nil
			}

			return nil
		},
	}
	return funs
}

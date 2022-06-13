package datatypes

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planKubespan NetworkKubeSpan) Data() (any, error) {
	kubespan := v1alpha1.NetworkKubeSpan{
		KubeSpanEnabled: planKubespan.Enabled.Value,
	}

	if !planKubespan.AllowPeerDownBypass.Null {
		kubespan.KubeSpanAllowDownPeerBypass = planKubespan.AllowPeerDownBypass.Value
	}

	return kubespan, nil
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
	*v1alpha1.NetworkKubeSpan
}

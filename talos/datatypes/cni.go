package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

// Data copies data from terraform state types to talos types.
func (cni CNI) Data() (any, error) {
	cniConfig := &v1alpha1.CNIConfig{}

	cniConfig.CNIName = cni.Name.Value
	for _, url := range cni.URLs {
		cniConfig.CNIUrls = append(cniConfig.CNIUrls, url.Value)
	}

	return cniConfig, nil
}

func (cniData *CNI) Read(cni any) error {
	cniConfig := cni.(*v1alpha1.CNIConfig)

	cniData.Name.Value = cniConfig.CNIName
	for _, url := range cniConfig.CNIUrls {
		cniData.URLs = append(cniData.URLs, types.String{Value: url})
	}

	return nil
}

func (cni CNI) GenOpts() (out []generate.GenOption, err error) {
	cniConfig, err := cni.Data()
	if err != nil {
		return nil, err
	}
	out = append(out, generate.WithClusterCNIConfig(cniConfig.(*v1alpha1.CNIConfig)))
	return
}

type TalosCNIConfig struct {
	*v1alpha1.CNIConfig
}

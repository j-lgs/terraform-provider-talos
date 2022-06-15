package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"gopkg.in/yaml.v3"
)

// Data copies data from terraform state types to talos types.
func (planAdmissionPluginConfig AdmissionPluginConfig) Data() (interface{}, error) {
	var admissionConfig v1alpha1.Unstructured

	if err := yaml.Unmarshal([]byte(planAdmissionPluginConfig.Configuration.Value), &admissionConfig); err != nil {
		return &v1alpha1.AdmissionPluginConfig{}, nil
	}

	admissionPluginConfig := &v1alpha1.AdmissionPluginConfig{
		PluginName:          planAdmissionPluginConfig.Name.Value,
		PluginConfiguration: admissionConfig,
	}
	return admissionPluginConfig, nil
}

type AdmissionControlConfigs []*v1alpha1.AdmissionPluginConfig
type TalosAdmissionPluginConfigs struct {
	AdmissionControlConfigs
}

func (talosAdmissionPluginConfigs TalosAdmissionPluginConfigs) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.APIServer.AdmissionPlugins == nil {
				planConfig.APIServer.AdmissionPlugins = make([]AdmissionPluginConfig, 0)
			}

			for _, config := range talosAdmissionPluginConfigs.AdmissionControlConfigs {
				conf := AdmissionPluginConfig{
					Name: readString(config.Name()),
				}

				var obj *types.String
				obj, err = readObject(config.PluginConfiguration)
				if err != nil {
					return
				}

				if obj != nil {
					conf.Configuration = *obj
				}

				planConfig.APIServer.AdmissionPlugins = append(planConfig.APIServer.AdmissionPlugins, conf)
			}

			return nil
		},
	}
	return funs
}

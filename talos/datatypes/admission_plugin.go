package datatypes

import (
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

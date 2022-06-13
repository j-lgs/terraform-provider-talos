package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"gopkg.in/yaml.v3"
)

// Data copies data from terraform state types to talos types.
func (planAPIServer APIServerConfig) Data() (interface{}, error) {
	apiServer := &v1alpha1.APIServerConfig{}

	if !planAPIServer.Image.Null {
		apiServer.ContainerImage = planAPIServer.Image.Value
	}
	apiServer.ExtraArgsConfig = map[string]string{}
	for arg, value := range planAPIServer.ExtraArgs {
		apiServer.ExtraArgsConfig[arg] = value.Value
	}
	if !planAPIServer.DisablePSP.Null {
		apiServer.DisablePodSecurityPolicyConfig = planAPIServer.DisablePSP.Value
	}

	for i, pluginYaml := range planAPIServer.AdmissionPlugins {
		apiServer.AdmissionControlConfig = append(apiServer.AdmissionControlConfig, &v1alpha1.AdmissionPluginConfig{
			PluginName: pluginYaml.Name.Value,
		})

		var plugin v1alpha1.Unstructured
		if err := yaml.Unmarshal([]byte(pluginYaml.Configuration.Value), &plugin); err != nil {
			return &v1alpha1.APIServerConfig{}, err
		}
		apiServer.AdmissionControlConfig[i].PluginConfiguration = plugin
	}

	for _, san := range planAPIServer.CertSANS {
		apiServer.CertSANs = append(apiServer.CertSANs, san.Value)
	}
	apiServer.EnvConfig = map[string]string{}
	for arg, value := range planAPIServer.Env {
		apiServer.EnvConfig[arg] = value.Value
	}
	for _, vol := range planAPIServer.ExtraVolumes {
		d, err := vol.Data()
		if err != nil {
			return &v1alpha1.APIServerConfig{}, err
		}
		apiServer.ExtraVolumesConfig = append(apiServer.ExtraVolumesConfig, d.(v1alpha1.VolumeMountConfig))
	}

	return apiServer, nil
}

func (planAPIServer APIServerConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			ins, err := planAPIServer.Data()
			if err != nil {
				return err
			}
			cfg.ClusterConfig.APIServerConfig = ins.(*v1alpha1.APIServerConfig)
			return nil
		},
	}
}

type TalosAPIServerConfig struct {
	*v1alpha1.APIServerConfig
}

func (talosAPIServerConfig TalosAPIServerConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.APIServer == nil {
				planConfig.APIServer = &APIServerConfig{}
			}

			planConfig.APIServer.Image = readString(talosAPIServerConfig.Image())
			planConfig.APIServer.DisablePSP = readBool(talosAPIServerConfig.DisablePodSecurityPolicyConfig)
			planConfig.APIServer.Env = readStringMap(talosAPIServerConfig.EnvConfig)
			planConfig.APIServer.ExtraArgs = readStringMap(talosAPIServerConfig.ExtraArgsConfig)

			return nil
		},
	}

	funs = append(funs, TalosKubeletMounts{VolumeMounts: talosAPIServerConfig.ExtraVolumesConfig}.ReadFunc()...)
	funs = append(funs, TalosAdmissionPluginConfigs{AdmissionControlConfigs: talosAPIServerConfig.AdmissionControlConfig}.ReadFunc()...)

	return funs
}

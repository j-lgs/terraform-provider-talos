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

	if len(planAPIServer.ExtraArgs) > 0 {
		apiServer.ExtraArgsConfig = map[string]string{}
		for arg, value := range planAPIServer.ExtraArgs {
			apiServer.ExtraArgsConfig[arg] = value.Value
		}
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

func (planAPIServer APIServerConfig) zero() bool {
	return mkString(planAPIServer.Image).zero() &&
		len(planAPIServer.ExtraArgs) <= 0 &&
		mkBool(planAPIServer.DisablePSP).zero() &&
		len(planAPIServer.AdmissionPlugins) <= 0 &&
		len(planAPIServer.CertSANS) <= 0 &&
		len(planAPIServer.Env) <= 0 &&
		len(planAPIServer.ExtraVolumes) <= 0
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

			if talosAPIServerConfig.ContainerImage != (&v1alpha1.APIServerConfig{}).Image() {
				planConfig.APIServer.Image = readString(talosAPIServerConfig.Image())
			}
			if planConfig.APIServer.Image.Value == "" {
				planConfig.APIServer.Image.Value = (&v1alpha1.APIServerConfig{}).Image()
			}

			planConfig.APIServer.DisablePSP = readBool(talosAPIServerConfig.DisablePodSecurityPolicyConfig)
			planConfig.APIServer.Env = readStringMap(talosAPIServerConfig.EnvConfig)

			if len(talosAPIServerConfig.ExtraArgsConfig) > 0 {
				planConfig.APIServer.ExtraArgs = readStringMap(talosAPIServerConfig.ExtraArgsConfig)
			}

			if len(talosAPIServerConfig.CertSANs) > 0 {
				planConfig.APIServer.CertSANS = readStringList(talosAPIServerConfig.CertSANs)
			}

			if planConfig.APIServer.zero() {
				planConfig.APIServer = nil
			}

			return nil
		},
	}

	if len(talosAPIServerConfig.ExtraVolumesConfig) > 0 {
		funs = append(funs, TalosAPIServerMounts{VolumeMounts: talosAPIServerConfig.ExtraVolumesConfig}.ReadFunc()...)
	}

	if len(talosAPIServerConfig.AdmissionControlConfig) > 0 {
		funs = append(funs, TalosAdmissionPluginConfigs{AdmissionControlConfigs: talosAPIServerConfig.AdmissionControlConfig}.ReadFunc()...)
	}

	return funs
}

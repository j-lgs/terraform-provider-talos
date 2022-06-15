package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planProxy ProxyConfig) Data() (any, error) {
	proxy := &v1alpha1.ProxyConfig{}
	if !planProxy.Image.Null {
		proxy.ContainerImage = planProxy.Image.Value
	}
	if !planProxy.Disabled.Null {
		proxy.Disabled = planProxy.Disabled.Value
	}
	if !planProxy.Mode.Null {
		proxy.ModeConfig = planProxy.Mode.Value
	}
	proxy.ExtraArgsConfig = map[string]string{}
	for arg, value := range planProxy.ExtraArgs {
		proxy.ExtraArgsConfig[arg] = value.Value
	}

	return proxy, nil
}

func (planProxy ProxyConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	funs := [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			proxy, err := planProxy.Data()
			if err != nil {
				return err
			}
			cfg.ClusterConfig.ProxyConfig = proxy.(*v1alpha1.ProxyConfig)

			return nil
		},
	}
	return funs
}

type TalosProxyConfig struct {
	*v1alpha1.ProxyConfig
}

func (proxyConfig ProxyConfig) zero() bool {
	return mkString(proxyConfig.Image).zero() &&
		mkBool(proxyConfig.Disabled).zero() &&
		len(proxyConfig.ExtraArgs) <= 0 &&
		mkString(proxyConfig.Mode).zero()
}

func (talosProxyConfig TalosProxyConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if talosProxyConfig.ProxyConfig == nil {
				return nil
			}

			if planConfig.Proxy == nil {
				planConfig.Proxy = &ProxyConfig{}
			}

			if talosProxyConfig.ContainerImage != (&v1alpha1.ProxyConfig{}).Image() {
				planConfig.Proxy.Image = readString(talosProxyConfig.Image())
			}
			if planConfig.Proxy.Image.Value == "" {
				planConfig.Proxy.Image.Value = (&v1alpha1.ProxyConfig{}).Image()
			}

			mkBool(talosProxyConfig.Disabled).read(&planConfig.Proxy.Disabled)

			readStringMap_(talosProxyConfig.ExtraArgsConfig, &planConfig.Proxy.ExtraArgs)

			mkString(talosProxyConfig.ModeConfig).read(&planConfig.Proxy.Mode)

			if planConfig.Proxy.zero() {
				planConfig.Proxy = nil
			}

			return nil
		},
	}
	return funs
}

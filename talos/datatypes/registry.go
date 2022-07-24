package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/crypto/x509"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

// Data copies data from terraform state types to talos types.
func (registry Registry) Data() (interface{}, error) {
	regs := &v1alpha1.RegistriesConfig{}

	regs.RegistryMirrors = map[string]*v1alpha1.RegistryMirrorConfig{}
	for registry, endpoints := range registry.Mirrors {
		regs.RegistryMirrors[registry] = &v1alpha1.RegistryMirrorConfig{}
		for _, endpoint := range endpoints {
			regs.RegistryMirrors[registry].MirrorEndpoints = append(regs.RegistryMirrors[registry].MirrorEndpoints, endpoint.Value)
		}
	}

	if registry.Configs != nil {
		regs.RegistryConfig = map[string]*v1alpha1.RegistryConfig{}
		for registry, conf := range registry.Configs {
			config, err := conf.Data()
			if err != nil {
				return nil, err
			}
			regs.RegistryConfig[registry] = config.(*v1alpha1.RegistryConfig)

		}
	}

	return regs, nil
}

// WithRegistryMirrors adds registry mirrors.
func WithRegistryMirrors(params map[string]*v1alpha1.RegistryMirrorConfig) generate.GenOption {
	return func(o *generate.GenOptions) error {
		o.RegistryMirrors = params
		return nil
	}
}

// WithRegistryConfigs add registry configurations.
func WithRegistryConfigs(params map[string]*v1alpha1.RegistryConfig) generate.GenOption {
	return func(o *generate.GenOptions) error {
		o.RegistryConfig = params
		return nil
	}
}

func (install Registry) GenOpts() (out []generate.GenOption, err error) {
	if install.Mirrors != nil {
		mirrors := map[string]*v1alpha1.RegistryMirrorConfig{}
		for host, mirrorlist := range install.Mirrors {
			mirrors[host] = &v1alpha1.RegistryMirrorConfig{}
			for _, mirror := range mirrorlist {
				mirrors[host].MirrorEndpoints = append(mirrors[host].MirrorEndpoints, mirror.Value)
			}
		}
		out = append(out, WithRegistryMirrors(mirrors))
	}

	if install.Configs != nil {
		configs := map[string]*v1alpha1.RegistryConfig{}
		for host, config := range install.Configs {
			conf, err := config.Data()
			if err != nil {
				return nil, err
			}
			configs[host] = conf.(*v1alpha1.RegistryConfig)
		}
		out = append(out, WithRegistryConfigs(configs))
	}

	return
}

type TalosRegistriesConfig struct {
	*v1alpha1.RegistriesConfig
}

func (talosRegistriesConfig TalosRegistriesConfig) Read() *Registry {
	r := Registry{
		Mirrors: map[string][]types.String{},
		Configs: map[string]RegistryConfig{},
	}

	if talosRegistriesConfig.Mirrors() != nil {
		for host, mirror := range talosRegistriesConfig.Mirrors() {
			for _, m := range mirror.Endpoints() {
				r.Mirrors[host] = append(r.Mirrors[host], types.String{Value: m})
			}
		}
	}

	rc := TalosRegistryConfigs{RegistryConfigs: talosRegistriesConfig.RegistryConfig}
	if talosRegistriesConfig.Config() != nil {
		r.Configs = rc.Read()
	}

	return &r
}

func (talosRegistriesConfig TalosRegistriesConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if talosRegistriesConfig.RegistriesConfig == nil {
				return nil
			}

			if planConfig.Registry == nil {
				planConfig.Registry = &Registry{}
			}

			if planConfig.Registry.Mirrors == nil {
				planConfig.Registry.Mirrors = make(map[string][]types.String)
			}

			for k, v := range talosRegistriesConfig.RegistryMirrors {
				planConfig.Registry.Mirrors[k] = readStringList(v.MirrorEndpoints)
			}

			return nil
		},
	}

	if len(talosRegistriesConfig.RegistryConfig) > 0 {
		funs = append(funs, TalosRegistryConfigs{RegistryConfigs: talosRegistriesConfig.RegistryConfig}.ReadFunc()...)
	}

	return funs
}

// Data copies data from terraform state types to talos types.
func (config RegistryConfig) Data() (interface{}, error) {
	conf := &v1alpha1.RegistryConfig{}

	conf.RegistryTLS = &v1alpha1.RegistryTLSConfig{}
	conf.RegistryAuth = &v1alpha1.RegistryAuthConfig{}
	if !config.ClientCRT.Null {
		conf.RegistryTLS.TLSClientIdentity = &x509.PEMEncodedCertificateAndKey{
			Crt: []byte(config.ClientCRT.Value),
			Key: []byte(config.ClientKey.Value),
		}
	}
	if !config.CA.Null {
		conf.RegistryTLS.TLSCA = []byte(config.CA.Value)
	}
	if !config.InsecureSkipVerify.Null {
		conf.RegistryTLS.TLSInsecureSkipVerify = config.InsecureSkipVerify.Value
	}

	if !config.Username.Null && !config.Password.Null {
		conf.RegistryAuth.RegistryUsername = config.Username.Value
		conf.RegistryAuth.RegistryPassword = config.Password.Value
	}

	if !config.Auth.Null && !config.IdentityToken.Null {
		conf.RegistryAuth.RegistryAuth = config.Auth.Value
		conf.RegistryAuth.RegistryIdentityToken = config.IdentityToken.Value
	}

	return conf, nil
}

func (registry Registry) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			reg, err := registry.Data()
			if err != nil {
				return err
			}
			cfg.MachineConfig.MachineRegistries = *reg.(*v1alpha1.RegistriesConfig)
			return nil
		},
	}
}

type RegistryConfigs map[string]*v1alpha1.RegistryConfig
type TalosRegistryConfigs struct {
	RegistryConfigs
}

func (talosRegistryConfigs TalosRegistryConfigs) Read() map[string]RegistryConfig {
	regs := make(map[string]RegistryConfig)

	for host, config := range talosRegistryConfigs.RegistryConfigs {
		registry := RegistryConfig{}

		if config.RegistryTLS != nil {
			registry.ClientCRT = readCert(*config.TLS().ClientIdentity())
			registry.ClientKey = readKey(*config.TLS().ClientIdentity())

			registry.CA = readString(string(config.TLS().CA()))
			registry.InsecureSkipVerify = readBool(config.TLS().InsecureSkipVerify())
		}

		if config.RegistryAuth != nil {
			registry.Username = readString(config.Auth().Username())
			registry.Password = readString(config.RegistryAuth.Password())

			registry.Auth = readString(config.Auth().Auth())
			registry.IdentityToken = readString(config.Auth().IdentityToken())
		}

		regs[host] = registry
	}

	return regs
}

func (talosRegistryConfigs TalosRegistryConfigs) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.Registry.Configs == nil {
				planConfig.Registry.Configs = make(map[string]RegistryConfig)
			}

			for host, config := range talosRegistryConfigs.RegistryConfigs {
				registry := RegistryConfig{}

				if config.RegistryTLS != nil {
					registry.ClientCRT = readCert(*config.TLS().ClientIdentity())
					registry.ClientKey = readKey(*config.TLS().ClientIdentity())

					registry.CA = readString(string(config.TLS().CA()))
					registry.InsecureSkipVerify = readBool(config.TLS().InsecureSkipVerify())
				}

				if config.RegistryAuth != nil {
					registry.Username = readString(config.Auth().Username())
					registry.Password = readString(config.RegistryAuth.Password())

					registry.Auth = readString(config.Auth().Auth())
					registry.IdentityToken = readString(config.Auth().IdentityToken())
				}

				planConfig.Registry.Configs[host] = registry
			}

			return nil
		},
	}
	return funs
}

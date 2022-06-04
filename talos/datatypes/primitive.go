package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"gopkg.in/yaml.v3"
)

type MachineSysfs map[string]types.String

func (sysfs MachineSysfs) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			val := map[string]string{}
			for key, str := range sysfs {
				val[key] = str.Value
			}

			cfg.MachineConfig.MachineSysfs = val
			return nil
		},
	}
}

type MachineSysctls map[string]types.String

func (sysfs MachineSysctls) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			/*
				val, err := any(sysfs).(DTStringMap).Data()
				if err != nil {
					return err
				}
				cfg.MachineConfig.MachineSysctls = val.(map[string]string)
			*/

			return nil
		},
	}
}

type MachineEnv map[string]types.String

func (sysfs MachineEnv) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			/*
				val, err := any(sysfs).(DTStringMap).Data()
				if err != nil {
					return err
				}
				cfg.MachineConfig.MachineSysctls = val.(map[string]string)
			*/
			return nil
		},
	}
}

type MachineCertSAN struct{ Value types.String }

func (certSAN MachineCertSAN) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			cfg.MachineConfig.MachineCertSANs = append(cfg.MachineConfig.MachineCertSANs, certSAN.Value.Value)
			return nil
		},
	}
}

type MachineUdev struct{ Value types.String }

func (rule MachineUdev) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			if cfg.MachineConfig.MachineUdev == nil {
				cfg.MachineConfig.MachineUdev = &v1alpha1.UdevConfig{}
			}
			cfg.MachineConfig.MachineUdev.UdevRules = append(cfg.MachineConfig.MachineUdev.UdevRules, rule.Value.Value)
			return nil
		},
	}
}

type MachinePod struct{ Value types.String }

func (pod MachinePod) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			var talosPod v1alpha1.Unstructured

			if err := yaml.Unmarshal([]byte(pod.Value.Value), &talosPod); err != nil {
				return err
			}
			cfg.MachineConfig.MachinePods = append(cfg.MachineConfig.MachinePods, talosPod)
			return nil
		},
	}
}

type ClusterExtraManifest struct{ Value types.String }

func (manifest ClusterExtraManifest) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			cfg.ClusterConfig.ExtraManifests = append(cfg.ClusterConfig.ExtraManifests, manifest.Value.Value)
			return nil
		},
	}
}

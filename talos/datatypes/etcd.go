package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planEtcd EtcdConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			etcd := cfg.ClusterConfig.EtcdConfig

			etcd.ContainerImage = (&v1alpha1.EtcdConfig{}).Image()
			if !planEtcd.Image.Null {
				etcd.ContainerImage = planEtcd.Image.Value

			}

			setCertKey(planEtcd.CaCrt, planEtcd.CaKey, etcd.RootCA)
			setStringMap(planEtcd.ExtraArgs, &etcd.EtcdExtraArgs)
			setString(planEtcd.Subnet, &etcd.EtcdSubnet)

			return nil
		},
	}
}

type TalosEtcdConfig struct {
	*v1alpha1.EtcdConfig
}

func (talosEtcdConfig TalosEtcdConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.Etcd == nil {
				planConfig.Etcd = &EtcdConfig{}
			}

			if talosEtcdConfig.ContainerImage != (&v1alpha1.EtcdConfig{}).Image() {
				planConfig.Etcd.Image = readString(talosEtcdConfig.Image())
			}
			if planConfig.Etcd.Image.Value == "" {
				planConfig.Etcd.Image.Value = (&v1alpha1.EtcdConfig{}).Image()
			}

			planConfig.Etcd.ExtraArgs = readStringMap(talosEtcdConfig.ExtraArgs())
			planConfig.Etcd.Subnet = readString(talosEtcdConfig.Subnet())
			planConfig.Etcd.CaKey = readKey(*talosEtcdConfig.CA())
			planConfig.Etcd.CaCrt = readCert(*talosEtcdConfig.CA())

			return
		},
	}
	return funs
}

package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func (planEtcd EtcdConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			etcd := cfg.ClusterConfig.EtcdConfig

			if planEtcd.Image.Null {
				etcd.ContainerImage = (&v1alpha1.EtcdConfig{}).Image()
			}
			setString(planEtcd.Image, &etcd.ContainerImage)
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

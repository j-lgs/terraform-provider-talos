package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

// Data copies data from terraform state types to talos types.
func (install InstallConfig) Data() (any, error) {
	installConfig := &v1alpha1.InstallConfig{}

	installConfig.InstallDisk = generate.DefaultGenOptions().InstallDisk
	setString(install.Disk, &installConfig.InstallDisk)

	mkString(install.Image).set(&installConfig.InstallImage)

	setBool(install.Wipe, &installConfig.InstallWipe)
	setBool(install.LegacyBios, &installConfig.InstallLegacyBIOSSupport)

	for _, extension := range install.Extensions {
		installConfig.InstallExtensions = append(installConfig.InstallExtensions, v1alpha1.InstallExtensionConfig{
			ExtensionImage: extension.Value,
		})
	}

	setStringList(install.KernelArgs, &installConfig.InstallExtraKernelArgs)

	setBool(install.Bootloader, &installConfig.InstallBootloader)

	return installConfig, nil
}

func (install InstallConfig) DataFunc() [](func(*v1alpha1.Config) error) {
	return [](func(*v1alpha1.Config) error){
		func(cfg *v1alpha1.Config) error {
			ins, err := install.Data()
			if err != nil {
				return err
			}
			cfg.MachineConfig.MachineInstall = ins.(*v1alpha1.InstallConfig)
			return nil
		},
	}
}

func (install InstallConfig) GenOpts() (out []generate.GenOption, err error) {
	if !install.Image.Null {
		out = append(out, generate.WithInstallImage(install.Image.Value))
	}

	if !install.Disk.Null {
		out = append(out, generate.WithInstallDisk(install.Disk.Value))
	}

	if len(install.KernelArgs) > 0 {
		kargs := []string{}
		for _, karg := range install.KernelArgs {
			kargs = append(kargs, karg.Value)
		}
		out = append(out, generate.WithInstallExtraKernelArgs(kargs))
	}
	return
}

type TalosInstallConfig struct {
	*v1alpha1.InstallConfig
}

func (talosInstallConfig TalosInstallConfig) ReadFunc() []ConfigReadFunc {
	funs := []ConfigReadFunc{
		func(planConfig *TalosConfig) (err error) {
			if planConfig.Install == nil {
				planConfig.Install = &InstallConfig{}
			}

			mkString(talosInstallConfig.InstallImage).read(&planConfig.Install.Image)

			mkBool(talosInstallConfig.InstallBootloader).read(&planConfig.Install.Bootloader)
			mkBool(talosInstallConfig.LegacyBIOSSupport()).read(&planConfig.Install.LegacyBios)
			mkBool(talosInstallConfig.InstallWipe).read(&planConfig.Install.Wipe)

			planConfig.Install.Disk = readString(talosInstallConfig.InstallDisk)
			planConfig.Install.KernelArgs = readStringList(talosInstallConfig.InstallExtraKernelArgs)

			if len(talosInstallConfig.InstallExtensions) > 0 {
				planConfig.Install.Extensions = make([]types.String, 0)
			}

			for _, extension := range talosInstallConfig.InstallExtensions {
				planConfig.Install.Extensions = append(planConfig.Install.Extensions, readString(extension.Image()))
			}

			return nil
		},
	}
	return funs
}

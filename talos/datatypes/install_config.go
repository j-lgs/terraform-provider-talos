package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

// Data copies data from terraform state types to talos types.
func (install InstallConfig) Data() (any, error) {
	installConfig := &v1alpha1.InstallConfig{}

	installConfig.InstallDisk = generate.DefaultGenOptions().InstallDisk
	if !install.Disk.Null {
		installConfig.InstallDisk = install.Disk.Value
	}

	installConfig.InstallImage = generate.DefaultGenOptions().InstallImage
	if !install.Image.Null {
		installConfig.InstallImage = install.Image.Value
	}

	for _, karg := range install.KernelArgs {
		installConfig.InstallExtraKernelArgs = append(installConfig.InstallExtraKernelArgs, karg.Value)
	}

	installConfig.InstallBootloader = true

	return installConfig, nil
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

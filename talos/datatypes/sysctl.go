package datatypes

import "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"

func (sysctls SysctlData) GenOpts() (out []generate.GenOption, err error) {
	data := map[string]string{}
	for key, value := range sysctls {
		data[key] = value.Value
	}

	return []generate.GenOption{generate.WithSysctls(data)}, nil
}

package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (mount ExtraMount) Data() (interface{}, error) {
	extraMount := v1alpha1.ExtraMount{
		Mount: specs.Mount{
			Destination: mount.Destination.Value,
			Source:      mount.Source.Value,
			Type:        mount.Type.Value,
		},
	}

	for _, opt := range mount.Options {
		extraMount.Options = append(extraMount.Options, opt.Value)
	}

	return extraMount, nil
}

// Read copies data from talos types to terraform state types.
func (mount *ExtraMount) Read(mnt interface{}) error {
	talosMount := mnt.(v1alpha1.ExtraMount)
	mount.Destination = types.String{Value: talosMount.Destination}
	mount.Source = types.String{Value: talosMount.Source}

	if talosMount.Type != "" {
		mount.Type = types.String{Value: talosMount.Type}
	}

	for _, opt := range talosMount.Options {
		mount.Options = append(mount.Options, types.String{Value: opt})
	}

	return nil
}

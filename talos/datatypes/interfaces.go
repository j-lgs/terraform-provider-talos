package datatypes

import "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"

// Interface defining methods used to move data to and from talos and terraform.
//lint:ignore U1000 type exists just to define the interface and isn't used by itself.
type PlanToAPI interface {
	Data() (interface{}, error)
	Read(interface{}) error
}

// For initial cluster wide configuration.
type PlanToGenopts interface {
	GenOpts() ([]generate.GenOption, error)
}

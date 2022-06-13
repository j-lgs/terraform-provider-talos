package datatypes

import (
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

// Interface defining methods used to move data to and from talos and terraform.
//lint:ignore U1000 type exists just to define the interface and isn't used by itself.
type PlanToAPI interface {
	Data() (any, error)
	Read(any) error
}

type ConfigDataFunc = func(*v1alpha1.Config) error

type PlanToDataFunc interface {
	DataFunc() []ConfigDataFunc
}

func AppendDataFuncs(in []PlanToDataFunc, funs []PlanToDataFunc) (out []PlanToDataFunc) {
	return append(in, funs...)
}

// For reading the node's data
type ConfigReadFunc = func(*TalosConfig) error

type ConfigToPlanFunc interface {
	ReadFunc() []ConfigReadFunc
}

func AppendReadFuncs(in []ConfigToPlanFunc, funs []ConfigToPlanFunc) (out []ConfigToPlanFunc) {
	return append(in, funs...)
}

// For initial cluster wide configuration.
type PlanToGenopts interface {
	GenOpts() ([]generate.GenOption, error)
}

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

func ToSliceOfAny[T any](s []T) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

func AppendDataFuncs(in []PlanToDataFunc, funs any) (out []PlanToDataFunc) {
	out = in
	for _, fun := range funs.([]any) {
		out = append(out, fun.(PlanToDataFunc))
	}
	return out
}

// For initial cluster wide configuration.
type PlanToGenopts interface {
	GenOpts() ([]generate.GenOption, error)
}

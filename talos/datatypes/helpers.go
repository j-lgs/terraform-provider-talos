package datatypes

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func s(str string) types.String {
	return types.String{Value: str}
}

func Wrapsl(strs ...string) (out []types.String) {
	for _, st := range strs {
		out = append(out, s(st))
	}
	return
}

func Wrapi(i int) types.Int64 {
	return types.Int64{Value: int64(i)}
}

func Wraps(s string) types.String {
	return types.String{Value: s}
}

func Wrapb(b bool) types.Bool {
	return types.Bool{Value: b}
}

func AppendDataFunc(in []ConfigDataFunc, readers ...PlanToDataFunc) (out []ConfigDataFunc) {
	out = in

	for _, data := range readers {
		if !reflect.ValueOf(data).IsZero() {
			out = append(out, data.DataFunc()...)
		}
	}

	return
}

func ApplyDataFunc(cfg *v1alpha1.Config, funcs []ConfigDataFunc) error {
	fmt.Printf("length %d\n", len(funcs))

	for _, f := range funcs {
		if err := f(cfg); err != nil {
			return err
		}
	}

	return nil
}

package datatypes

import "github.com/hashicorp/terraform-plugin-framework/types"

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

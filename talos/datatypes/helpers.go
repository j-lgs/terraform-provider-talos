package datatypes

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/crypto/x509"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"

	"gopkg.in/yaml.v3"
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
	for _, f := range funcs {
		if err := f(cfg); err != nil {
			return err
		}
	}

	return nil
}

func setBool(val types.Bool, dest *bool) {
	if val.Null {
		return
	}
	*dest = val.Value
}

func setString(val types.String, dest *string) {
	if val.Null {
		return
	}

	*dest = val.Value
}

func setStringDuration(str types.String, dest *time.Duration) error {
	if str.Null {
		return nil
	}

	dur, err := time.ParseDuration(str.Value)
	if err != nil {
		return err
	}

	*dest = dur

	return nil
}

func setEndpoint(str types.String, dest *v1alpha1.Endpoint) error {

	return nil
}

func setCertKey(crt types.String, key types.String, dest *x509.PEMEncodedCertificateAndKey) {
	if crt.Null || key.Null {
		return
	}

	dest = &x509.PEMEncodedCertificateAndKey{
		Crt: []byte(crt.Value),
		Key: []byte(key.Value),
	}
}

// TODO return errors whenever a null pointer is passed
func setStringList(list []types.String, dest *[]string) {
	if dest == nil {
		return // fmt.Errorf("null destination provided")
	}

	if len(list) <= 0 {
		return
	}

	if len(*dest) == 0 {
		*dest = []string{}
	}

	for _, s := range list {
		*dest = append(*dest, s.Value)
	}
}

func setStringMap(valmap map[string]types.String, dest *map[string]string) {
	if dest == nil {
		return // fmt.Errorf("null destination provided")
	}

	if len(valmap) <= 0 {
		return
	}

	if len(*dest) == 0 {
		*dest = map[string]string{}
	}

	for k, v := range valmap {
		(*dest)[k] = v.Value
	}
}

func setVolumeMounts(mounts []VolumeMount, dest *[]v1alpha1.VolumeMountConfig) error {
	if dest == nil {
		return fmt.Errorf("null destination provided")
	}

	if len(mounts) <= 0 {
		return nil
	}

	if len(*dest) == 0 {
		*dest = []v1alpha1.VolumeMountConfig{}
	}

	for _, mount := range mounts {
		m, err := mount.Data()
		if err != nil {
			return err
		}
		*dest = append((*dest), m.(v1alpha1.VolumeMountConfig))
	}

	return nil
}

func setObjectList(yamls []types.String, dest *[]v1alpha1.Unstructured) error {
	if dest == nil {
		return fmt.Errorf("null destination provided")
	}

	if len(yamls) <= 0 {
		return nil
	}

	if len(*dest) == 0 {
		*dest = []v1alpha1.Unstructured{}
	}

	for _, yamlObject := range yamls {
		var object v1alpha1.Unstructured
		if err := yaml.Unmarshal([]byte(yamlObject.Value), &object); err != nil {
			return err
		}
		*dest = append(*dest, object)
	}

	return nil
}

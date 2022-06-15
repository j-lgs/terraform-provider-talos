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

func AppendReadFunc(in []ConfigReadFunc, readers ...ConfigToPlanFunc) (out []ConfigReadFunc) {
	out = in

	for _, data := range readers {
		if !reflect.ValueOf(data).IsZero() {
			out = append(out, data.ReadFunc()...)
		}
	}

	return
}

func ApplyReadFunc(talosConfig *TalosConfig, funcs []ConfigReadFunc) (TalosConfig, error) {
	for _, f := range funcs {
		if err := f(talosConfig); err != nil {
			return TalosConfig{}, err
		}
	}

	return *talosConfig, nil
}

type (
	StringType interface {
		read(*types.String)
		set(*string)
		zero() bool
	}

	StringTypeValue string
	StringTypeNone  struct{}
)

func mkString(val any) StringType {
	switch v := val.(type) {
	case string:
		if reflect.ValueOf(v).IsZero() {
			return StringTypeNone{}
		}
		return StringTypeValue(v)
	case types.String:
		if v.Null || reflect.ValueOf(v.Value).IsZero() {
			return StringTypeNone{}
		}
		return StringTypeValue(v.Value)
	default:
		panic("unsupported type passed to mkString")
	}
}

func (val StringTypeValue) read(out *types.String) {
	if out == nil {
		return
	}

	(*out) = types.String{Value: string(val)}
}
func (val StringTypeValue) set(out *string) {
	if out == nil {
		return
	}

	(*out) = string(val)
}
func (StringTypeValue) zero() bool {
	return false
}

func (StringTypeNone) read(out *types.String) {}
func (StringTypeNone) set(*string)            {}
func (StringTypeNone) zero() bool             { return true }

type (
	BoolType interface {
		read(*types.Bool)
		set(*bool)
		zero() bool
	}

	BoolTypeValue bool
	BoolTypeNone  struct{}
)

func mkBool(val any) BoolType {
	switch v := val.(type) {
	case bool:
		if !v {
			return BoolTypeNone{}
		}
		return BoolTypeValue(v)
	case types.Bool:
		if v.Null || !v.Value {
			return BoolTypeNone{}
		}

		return BoolTypeValue(v.Value)
	default:
		panic("unsupported type passed to mkBool")
	}
}

func (val BoolTypeValue) read(out *types.Bool) {
	if out == nil {
		return
	}

	(*out) = types.Bool{Value: bool(val)}
}

func (val BoolTypeValue) set(out *bool) {
	if out == nil {
		return
	}

	(*out) = bool(val)
}
func (BoolTypeValue) zero() bool { return false }

func (BoolTypeNone) read(out *types.Bool) {}

func (BoolTypeNone) set(*bool)  {}
func (BoolTypeNone) zero() bool { return true }

func readInt(val int) types.Int64 {
	return types.Int64{Value: int64(val)}
}

func setBool(val types.Bool, dest *bool) {
	if val.Null {
		return
	}

	*dest = val.Value
}

func readBool(val bool) types.Bool {
	return types.Bool{Value: val}
}

func setString(val types.String, dest *string) {
	if val.Null {
		return
	}

	*dest = val.Value
}

func readString(val string) types.String {
	return types.String{Value: val}
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

func readStringDuration(lifetime time.Duration) types.String {
	return types.String{Value: lifetime.String()}
}

func setEndpoint(str types.String, dest *v1alpha1.Endpoint) error {

	return nil
}

func readEndpoint(endpoint *url.URL) types.String {
	if reflect.ValueOf(endpoint).IsZero() {
		return types.String{Null: true}
	}

	return readString(endpoint.String())
}

func setCertKey(crt types.String, key types.String, dest *x509.PEMEncodedCertificateAndKey) {
	if crt.Null || key.Null {
		return
	}

	*dest = x509.PEMEncodedCertificateAndKey{
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

	for _, s := range list {
		*dest = append(*dest, s.Value)
	}
}

func readStringList(list []string) (dest []types.String) {
	if len(list) <= 0 {
		return
	}

	for _, s := range list {
		dest = append(dest, types.String{Value: s})
	}

	return
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

func readStringMap(valmap map[string]string) (dest map[string]types.String) {
	if len(valmap) <= 0 {
		return
	}

	dest = make(map[string]types.String)
	for k, v := range valmap {
		dest[k] = types.String{Value: v}
	}

	return
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

func readObject(object v1alpha1.Unstructured) (dest *types.String, err error) {
	bytes, err := yaml.Marshal(&object)
	if err != nil {
		return
	}

	if string(bytes) == `{}
` {
		return
	}

	s := readString(string(bytes))
	dest = &s

	return
}

func readObjectList(objects []v1alpha1.Unstructured) (dest []types.String, err error) {
	for _, object := range objects {
		var obj *types.String

		obj, err = readObject(object)
		if err != nil {
			return
		}

		if obj != nil {
			dest = append(dest, *obj)
		}
	}

	return
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

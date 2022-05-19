package talos

import (
	"context"
	"fmt"
	"net"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Schema validation helpers

/*
	ValidateFunc: func(value interface{}, key string) (warns []string, errs []error) {
		v := value.(string)
		config := kubeval.NewDefaultConfig()
		schemaCache := kubeval.NewSchemaCache()
		_, err := kubeval.ValidateWithCache([]byte(v), schemaCache, config)
		if err != nil {
			errs = append(errs, fmt.Errorf("Invalid kubernetes manifest provided"))
			errs = append(errs, err)
		}
		return
	},
*/

// validateDomain checks whether the provided schema value is a valid domain name.
func validateDomain(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain, got \"%s\"", key, v))
	}
	return
}

// validateDomain checks whether the provided schema value is a valid MAC address.
func validateMAC(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, err := net.ParseMAC(v); err != nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid MAC address, got \"%s\", error \"%s\"", key, v, err.Error()))
	}
	return
}

// validateDomain checks whether the provided schema value is a valid MAC address.
func validateCIDR(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, _, err := net.ParseCIDR(v); err != nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid CIDR IP address, got \"%s\", error \"%s\"", key, v, err.Error()))
	}
	return
}

// validateIP checks whether the provided schema value is a valid IP address.
func validateIP(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if net.ParseIP(v) == nil {
		errs = append(errs, fmt.Errorf("%s: Must provide a valid IP address, got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid hostname.
func validateHost(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*:[0-9]{2,}`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain with a port appended, seperated by \":\", got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid wireguard endpoint.
func validateEndpoint(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(:[0-9]{2,})?`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a lowercase RFC 1123 subdomain with an optional port appended, seperated by \":\", got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid image registry identifier. Used for specifying images for containers in Static Pods.
func validateImage(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	pattern := `[^/]+\.[^/.]+/([^/.]+/)?[^/.]+(:.+)?`
	if !regexp.MustCompile(pattern).Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%s: Node name must be a valid container image, got \"%s\"", key, v))
	}
	return
}

// validateIP checks whether the provided schema value is a valid wireguard public or private key.
func validateKey(value interface{}, key string) (warns []string, errs []error) {
	v := value.(string)
	if _, err := wgtypes.ParseKey(v); err != nil {
		errs = append(errs, err)
	}
	return
}

func StringListSchema(desc string) schema.Schema {
	return schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: desc,
		Default:     []string{},
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func StringListSchemaValidate(desc string, validate schema.SchemaValidateFunc) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: desc,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validate,
		},
	}
}

func listSuppressor(key string, old string, new string, d *schema.ResourceData) bool {

	return true
}

// Suppresses output according to the provided suppressListFunc
func Suppress(suppress schema.SchemaDiffSuppressFunc, in *schema.Schema) (out *schema.Schema) {
	out = in
	out.DiffSuppressFunc = suppress
	return
}

// Validate is a helper for building terraform resource schemas that denotes the passed schema is Required.
func Required(in *schema.Schema) (out *schema.Schema) {
	out = in
	out.Required = true
	return
}

// Validate is a helper for building terraform resource schemas that denotes the passed schema is Optional.
func Optional(in *schema.Schema) (out *schema.Schema) {
	out = in
	out.Optional = true
	return
}

// Validate is a helper for building terraform resource schemas that adds validation to the passed schema.
func Validate(in *schema.Schema, validate schema.SchemaValidateFunc) (out *schema.Schema) {
	out = in
	out.ValidateFunc = validate
	return
}

// StringMap is a helper for building terraform resource schemas that adds validation to the elements to the passed schema.
func ValidateInner(in *schema.Schema, validate schema.SchemaValidateFunc) (out *schema.Schema) {
	out = in
	out.Elem.(*schema.Schema).ValidateFunc = validate
	return
}

func IsValidPath(value interface{}) (ok bool, err error) {
	v := value.(string)
	// A valid filesystem path must not contain spaces
	if ok = !regexp.MustCompile(`\s`).Match([]byte(v)); !ok {
		err = fmt.Errorf("%s: Invalid mount path, must not contain spaces.", v)
	}
	return
}

type talosAttributeValidator struct {
	description         string
	markdownDescription string
	validate            func(context.Context, tfsdk.ValidateAttributeRequest, *tfsdk.ValidateAttributeResponse)
}

type TalosAttributeValidator interface {
	tfsdk.AttributeValidator
}

func (v *talosAttributeValidator) Description(context.Context) string {
	return v.description
}

func (v *talosAttributeValidator) MarkdownDescription(context.Context) string {
	return v.markdownDescription
}

func (v *talosAttributeValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	v.validate(ctx, req, resp)
}

// AllElemsValid takes a function that returns false whenever the field fails validation and returns a function that validates a ListType.
func AllElemsValid(f func(interface{}) (bool, error)) TalosAttributeValidator {
	val := new(talosAttributeValidator)
	val.description = "AllElemsValid takes a function that returns false whenever the field fails validation uses it to check all fields of a ListType."
	val.markdownDescription = val.description
	val.validate = func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
		l, err := req.AttributeConfig.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Unable to transform given typeList %s to a terraform value.", req.AttributePath.String()), err.Error())
			return
		}
		list := []interface{}{}
		l.As(list)

		for _, v := range list {
			if ok, err := f(v); !ok {
				resp.Diagnostics.AddAttributeError(req.AttributePath, "List object failed validation, error from validator function shown below.", err.Error())
				return
			}
		}
	}
	return val
}

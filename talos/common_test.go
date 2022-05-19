package talos

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	talosx509 "github.com/talos-systems/crypto/x509"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func TestPrepRegistryConfig(t *testing.T) {
	r := RegistryConfig{
		Username: types.String{Value: "test"},
		Password: types.String{Value: "test"},
	}
	expect := &v1alpha1.RegistryConfig{}
	expect.RegistryTLS = &v1alpha1.RegistryTLSConfig{}
	expect.RegistryTLS.TLSClientIdentity = &talosx509.PEMEncodedCertificateAndKey{}
	expect.RegistryAuth = &v1alpha1.RegistryAuthConfig{
		RegistryUsername: "test",
		RegistryPassword: "test",
	}

	got := prepRegistryConfig(r).(*v1alpha1.RegistryConfig)
	if !reflect.DeepEqual(*expect, *got) {
		t.Fatalf("Error matching output and expected: %#v - %#v", expect.RegistryAuth, got.RegistryAuth)
	}
}

package talos

import (
	"encoding/json"
	"reflect"
	"terraform-provider-talos/talos/datatypes"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/wI2L/jsondiff"
)

var (
	expectedInput = datatypes.InputBundleExample

	tfinput = talosClusterConfigResourceData{
		// For Resource
		TargetVersion:            datatypes.Wraps(""),
		Debug:                    datatypes.Wrapb(datatypes.ConfigDebugExample),
		AllowSchedulingOnMasters: datatypes.Wrapb(datatypes.AllowSchedulingOnMastersExample),
		ClusterName:              datatypes.Wraps(datatypes.ClusterNameExample),
		// Talos related
		MachineCertSANs:    []types.String{},
		K8sCertSANs:        []types.String{{Value: "1.2.3.4"}, {Value: "4.5.6.7"}},
		Install:            datatypes.InstallExampleClusterConfig,
		CNI:                datatypes.CniExample,
		Registry:           datatypes.RegistryExample,
		Sysctls:            datatypes.SysctlData(nodeData.Sysctls),
		Disks:              datatypes.TalosConfigExample.Disks,
		Encryption:         datatypes.EncryptionDataExample,
		KubernetesEndpoint: datatypes.Wraps(datatypes.EndpointExample.String()),
		KubernetesVersion:  datatypes.Wraps(testKubernetesVersion),
	}
)

func TestCreateTalosConfiguration(t *testing.T) {
	data := tfinput

	kubernetesVersion := data.KubernetesVersion.Value
	clusterName := data.ClusterName.Value

	secrets := &testBundle

	genopts, err := data.TalosData()
	if err != nil {
		t.Fatal(err)
	}
	input, err := generate.NewInput(clusterName, data.KubernetesEndpoint.Value, kubernetesVersion, secrets,
		genopts...,
	)
	if err != nil {
		t.Fatal(err)
	}

	testExpected := expectedInput
	// Test data for either needs to be slightly different, as the Talos generate function
	// for assigning Cert SANs changes both AdditionalMachineCertSANs and AdditionalSubjectAltNames.
	testExpected.AdditionalMachineCertSANs = testExpected.AdditionalSubjectAltNames
	testInput := *input

	testExpected.NetworkConfigOptions = nil
	if testInput.NetworkConfigOptions != nil {
		testInput.NetworkConfigOptions = nil
	}

	if !reflect.DeepEqual(testExpected, testInput) {
		inputJSON, err := json.MarshalIndent(testInput, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		expectedJSON, err := json.MarshalIndent(testExpected, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		patch, err := jsondiff.CompareJSON(expectedJSON, inputJSON)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("expected and actual input structures did not match\nchangelog %s", patch)
	}
}

func TestReadTalosConfiguration(t *testing.T) {
	var state = tfinput

	state.ReadInto(&expectedInput)

	if !reflect.DeepEqual(tfinput, state) {
		inputJSON, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		expectedJSON, err := json.MarshalIndent(tfinput, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		patch, err := jsondiff.CompareJSON(expectedJSON, inputJSON)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("expected and actual state did not match\nchangelog %s", patch)
	}
}

package talos

import (
	"encoding/json"
	"reflect"
	"terraform-provider-talos/talos/datatypes"
	"testing"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"github.com/wI2L/jsondiff"
)

var (
	expectedInput = datatypes.InputBundleExample

	tfinput = talosClusterConfigResourceData{
		// For Resource
		TargetVersion: datatypes.Wraps("v1.0.0"),
		ClusterName:   datatypes.Wraps(datatypes.ClusterNameExample),
		Endpoints:     datatypes.Wrapsl("0.0.0.0"),
		// Talos related
		Install:            datatypes.InstallExample,
		CNI:                datatypes.CniExample,
		Sysctls:            nodeData.Sysctls,
		Registry:           nodeData.Registry,
		Disks:              nodeData.Disks,
		Encryption:         nodeData.Encryption,
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
	var state talosClusterConfigResourceData = tfinput
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

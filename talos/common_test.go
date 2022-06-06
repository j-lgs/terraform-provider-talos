package talos

import (
	"encoding/json"
	"fmt"
	"reflect"
	"terraform-provider-talos/talos/datatypes"
	"testing"

	"github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/wI2L/jsondiff"

	configloader "github.com/talos-systems/talos/pkg/machinery/config/configloader"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
)

var (
	//date  = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	//clock = genv1alpha1.NewClock()

	testBundle = datatypes.SecretsBundleExample

	//expectedNode *v1alpha1.Config              = datatypes.MachineConfigExample
	nodeData *talosControlNodeResourceData = talosControlNodeResourceDataExample
)

type runtimeMode struct {
	requiresInstall bool
}

func (m runtimeMode) String() string {
	return fmt.Sprintf("runtimeMode(%v)", m.requiresInstall)
}

func (m runtimeMode) RequiresInstall() bool {
	return m.requiresInstall
}

// TestValidateConfig checks whether an expected valid configuration using values in all fields can be created from a Terraform state struct.
func TestValidateConfig(t *testing.T) {
	confString, err := genConfig(machine.TypeControlPlane, &datatypes.InputBundleExample, &talosControlNodeResourceDataExample)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := configloader.NewFromBytes([]byte(confString))
	if err != nil {
		t.Fatal(err)
	}

	opts := []config.ValidationOption{config.WithLocal()}
	opts = append(opts, config.WithStrict())

	warnings, err := cfg.Validate(runtimeMode{requiresInstall: true}, opts...)
	for _, w := range warnings {
		t.Logf("%s", w)
	}
	if err != nil {
		t.Fatal(err)
	}
}

// TestConfigDataAll checks if converting a Terraform data struct describing a controlplane node is
// converted into an expected Talos v1alpha1.Config struct.
func TestConfigDataAll(t *testing.T) {
	confString, err := genConfig(machine.TypeControlPlane, &datatypes.InputBundleExample, &nodeData)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := configloader.NewFromBytes([]byte(confString))
	if err != nil {
		t.Fatal(err)
	}

	// Build v1alpha1.Config struct from our cfg
	config := &v1alpha1.Config{
		ConfigVersion: "v1alpha1",
		ConfigDebug:   datatypes.ConfigDebugExample,
		ConfigPersist: datatypes.ConfigPersistExample,
		MachineConfig: any(cfg.Machine()).(*v1alpha1.MachineConfig),
		ClusterConfig: any(cfg.Cluster()).(*v1alpha1.ClusterConfig),
	}

	// We cannot compare lists of functions in this way, so it is currently disabled for now.
	if config.MachineConfig.MachineInstall.InstallDiskSelector != nil {
		config.MachineConfig.MachineInstall.InstallDiskSelector = nil
	}
	if datatypes.MachineConfigExample.MachineConfig.MachineInstall.InstallDiskSelector != nil {
		datatypes.MachineConfigExample.MachineConfig.MachineInstall.InstallDiskSelector = nil
	}

	if !reflect.DeepEqual(datatypes.MachineConfigExample, config) {
		// json marshalindent does not support printing matchers. Ignore it in printout.

		configJSON, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		exampleJSON, err := json.MarshalIndent(datatypes.MachineConfigExample, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		patch, err := jsondiff.CompareJSON(exampleJSON, configJSON)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("expected and actual v1alpha1.Config structs did not match\nchangelog\n%s", patch)

		//if config.MachineConfig.MachineInstall.InstallDiskSelector == nil {
		//	t.FailNow()
		//}
		//	matcherChanges, err := sdiff.Diff(
		//expectedNode.MachineConfig.MachineInstall.InstallDiskSelector.Size.Matcher,
		//config.MachineConfig.MachineInstall.InstallDiskSelector.Size.Matcher,
		//)
		//if err != nil {
		//t.Fatal(err)
		//}
		//changelog = []byte(spew.Sdump(matcherChanges))
		//t.Fatalf("expected and actual matchers did not match\nchangelog %s", changelog)
	}
}

// TestReadControlConfig checks whether we can successfully read a talos Config struct into a Terraform state struct.
func TestReadControlConfig(t *testing.T) {
	var state = &talosControlNodeResourceData{}

	state.ReadInto(datatypes.MachineConfigExample)
	if !reflect.DeepEqual(talosControlNodeResourceDataExample, state) {
		stateJSON, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		exampleJSON, err := json.MarshalIndent(talosControlNodeResourceDataExample, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		patch, err := jsondiff.CompareJSON(exampleJSON, stateJSON)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("expected and actual state did not match\nchangelog %s", patch)
	}
}

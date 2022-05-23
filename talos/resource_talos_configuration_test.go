package talos

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sdiff "github.com/r3labs/diff/v3"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
)

var (
	expectedInput = generate.Input{
		Certs:                      testDefaultCerts,
		VersionContract:            testVersionContract,
		ControlPlaneEndpoint:       testControlPlaneEndpoint.String(),
		AdditionalSubjectAltNames:  testSANs,
		AdditionalMachineCertSANs:  testCertSANs,
		ClusterID:                  testClusterID,
		ClusterName:                testClusterName,
		ClusterSecret:              testClusterSecret,
		ServiceDomain:              testServiceDomain,
		PodNet:                     testPodNet,
		ServiceNet:                 testServiceNet,
		KubernetesVersion:          testKubernetesVersion,
		Secrets:                    testSecrets,
		TrustdInfo:                 testTrustdInfo,
		ExternalEtcd:               testExternalEtcd,
		InstallDisk:                testInstallDisk,
		InstallImage:               testInstallImage,
		InstallExtraKernelArgs:     testInstallExtraKernelArgs,
		NetworkConfigOptions:       testNetworkConfigOptions,
		CNIConfig:                  testCNIConfig,
		RegistryMirrors:            testRegistryMirror,
		RegistryConfig:             testRegistryConfig,
		MachineDisks:               testMachineDisks,
		SystemDiskEncryptionConfig: testSystemDiskEncryptionConfig,
		Sysctls:                    testSysctls,
		Debug:                      testDebug,
		Persist:                    testPersist,
		AllowSchedulingOnMasters:   testAllowSchedulingOnMasters,
		DiscoveryEnabled:           testDiscoveryEnabled,
	}

	tfinput = talosClusterConfigResourceData{
		// For Resource
		TargetVersion: wraps("v1.0.0"),
		ClusterName:   wraps(testClusterName),
		Endpoints:     sl("0.0.0.0"),
		// Talos related
		Encryption: &EncryptionData{
			State: &EncryptionConfigData{
				Provider: wraps(testSystemDiskEncryptionConfig.StatePartition.EncryptionProvider),
				Keys: []KeyConfig{
					{
						NodeID:    types.Bool{Null: true},
						KeyStatic: wraps(testSystemDiskEncryptionConfig.StatePartition.EncryptionKeys[0].KeyStatic.KeyData),
						Slot:      wrapi(testSystemDiskEncryptionConfig.StatePartition.EncryptionKeys[0].KeySlot),
					},
				},
				Cipher:      wraps(testSystemDiskEncryptionConfig.StatePartition.EncryptionCipher),
				KeySize:     wrapi(int(testSystemDiskEncryptionConfig.StatePartition.EncryptionKeySize)),
				BlockSize:   wrapi(int(testSystemDiskEncryptionConfig.StatePartition.EncryptionBlockSize)),
				PerfOptions: sl(testSystemDiskEncryptionConfig.StatePartition.EncryptionPerfOptions...),
			},
			Ephemeral: &EncryptionConfigData{
				Provider: wraps(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionProvider),
				Keys: []KeyConfig{
					{
						KeyStatic: types.String{Null: true},
						NodeID:    wrapb(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionKeys[0].KeyNodeID != nil),
						Slot:      wrapi(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionKeys[0].KeySlot),
					},
				},
				Cipher:      wraps(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionCipher),
				KeySize:     wrapi(int(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionKeySize)),
				BlockSize:   wrapi(int(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionBlockSize)),
				PerfOptions: sl(testSystemDiskEncryptionConfig.EphemeralPartition.EncryptionPerfOptions...),
			},
		},
		KubernetesEndpoint: wraps(testControlPlaneEndpoint.String()),
		KubernetesVersion:  wraps(testKubernetesVersion),
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
	_, err = json.MarshalIndent(testExpected, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	changes, err := sdiff.Diff(testInput, testExpected)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) > 0 {
		changelog, err := json.MarshalIndent(changes, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("expected and actual input structures did not match\nchangelog %s", changelog)
	}
}

func TestReadTalosConfiguration(t *testing.T) {
	var state talosClusterConfigResourceData = tfinput
	testExpected := expectedInput

	state.ReadInto(&testExpected)
	changes, err := sdiff.Diff(state, tfinput)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) > 0 {
		changelog, err := json.MarshalIndent(changes, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("expected and actual state did not match\nchangelog %s", changelog)
	}
}

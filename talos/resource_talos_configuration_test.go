package talos

import (
	"encoding/json"
	"testing"

	sdiff "github.com/r3labs/diff/v3"
	"github.com/talos-systems/talos/pkg/machinery/config"
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
		TargetVersion:      wraps("v1.0.0"),
		ClusterName:        wraps(testClusterName),
		Endpoints:          sl("0.0.0.0"),
		KubernetesEndpoint: wraps(testControlPlaneEndpoint.String()),
		KubernetesVersion:  wraps(testKubernetesVersion),
	}
)

func TestCreateTalosConfiguration(t *testing.T) {
	data := tfinput

	targetVersion := data.TargetVersion.Value
	kubernetesVersion := data.KubernetesVersion.Value
	clusterName := data.ClusterName.Value

	endpoints := []string{}
	for _, e := range data.Endpoints {
		endpoints = append(endpoints, e.Value)
	}

	versionContract, err := config.ParseContractFromVersion(targetVersion)
	if err != nil {
		t.Fatal(err)
	}

	clock.SetFixedTimestamp(date)

	secrets := &testBundle

	input, err := generate.NewInput(clusterName, data.KubernetesEndpoint.Value, kubernetesVersion, secrets,
		generate.WithVersionContract(versionContract),
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
	var state talosClusterConfigResourceData
	//testExpected := expectedInput

	//state.ReadInto(testExpected)
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

package talos

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Global variables
var (
	testControlIPs   []string      = []string{"192.168.124.5", "192.168.124.6", "192.168.124.7"}
	testControlWGIPs []string      = []string{"192.168.125.5", "192.168.125.6", "192.168.125.7"}
	resetWaitTime    time.Duration = 10 * time.Second
)

// testResetVM is a workaround for Virtual machines hanging on reboot when reset by Talos API.
// It is toggled by the environment variable RESET_VM
func testResetVM(t *testing.T, vmprefix string, vmIndicies ...int) {
	// Enough time for the reset process to run to completion on a fast system.
	time.Sleep(resetWaitTime)
	for _, idx := range vmIndicies {
		var stdout bytes.Buffer
		cmd := exec.Command("virsh", "-c", "qemu:///system", "reset", "test_"+vmprefix+"_"+strconv.Itoa(idx))
		cmd.Stdout = &stdout
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

// TestAccResourceTalosControlSingleMaster runs tests involving a single master node.
func TestAccResourceTalosControlSingleMaster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"local": {
				VersionConstraint: "2.2.3",
				Source:            "hashicorp/local",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testTalosConfig(&testConfig{
					Endpoint: testControlIPs[0],
				}) + testControlConfig(&testNode{
					IP:        testControlIPs[0],
					MAC:       testMACAddresses[0],
					Index:     0,
					Bootstrap: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccTalosConnectivity(testControlNodePath(0), testControlIPs[0]),
					testAccKubernetesConnectivity("https://"+testControlIPs[0]+":6443"),
				),
			},
		},
	})
}

func TestAccResourceTalosControlThreeMaster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if _, reset := os.LookupEnv("RESET_VM"); reset {
				testResetVM(t, "control", 0, 1, 2)
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"local": {
				VersionConstraint: "2.2.3",
				Source:            "hashicorp/local",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testTalosConfig(&testConfig{
					Endpoint: testControlIPs[1],
				}) + testControlConfig(
					&testNode{
						IP:        testControlIPs[0],
						MAC:       testMACAddresses[0],
						Index:     0,
						Bootstrap: false,
					}, &testNode{
						IP:        testControlIPs[1],
						MAC:       testMACAddresses[1],
						Index:     1,
						Bootstrap: true,
					}, &testNode{
						IP:        testControlIPs[2],
						MAC:       testMACAddresses[2],
						Index:     2,
						Bootstrap: false,
					}),
				Check: resource.ComposeTestCheckFunc(
					testAccTalosConnectivity(testControlNodePath(1), testControlIPs[1]),
					testAccKubernetesConnectivity("https://"+testControlIPs[1]+":6443"),
				),
			},
		},
	})
}

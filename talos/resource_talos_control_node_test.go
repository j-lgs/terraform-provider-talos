package talos

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Global variables
var (
	testControlIPs []string = []string{"10.0.2.200", "10.0.2.201", "10.0.2.202"}
	// testControlWGIPs []string      = []string{"192.168.125.5", "192.168.125.6", "192.168.125.7"}
	resetWaitTime time.Duration = 3 * time.Second
)

// TestAccResourceTalosControlSingleMaster runs tests involving a single master node.
func TestAccResourceTalosControlSingleMaster(t *testing.T) {
	ips := []string{}
	for current := testInitialIPs.From(); current != testInitialIPs.To(); current = current.Next() {
		ips = append(ips, current.String())
	}

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
					IP:          testControlIPs[0],
					ProvisionIP: ips[0],
					Index:       0,
					Bootstrap:   true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("talos_control_node.control_0", "name"),
					resource.TestCheckResourceAttrSet("talos_control_node.control_0", "bootstrap"),
					resource.TestCheckResourceAttrSet("talos_control_node.control_0", "configure_ip"),
					resource.TestCheckResourceAttrSet("talos_control_node.control_0", "provision_ip"),
					resource.TestCheckResourceAttrSet("talos_control_node.control_0", "install.disk"),
					resource.TestCheckResourceAttrSet("talos_control_node.control_0", "base_config"),

					resource.TestCheckResourceAttr("talos_control_node.control_0",
						"networkconfig.devices.0.name", "eth0"),
					resource.TestCheckResourceAttr("talos_control_node.control_0",
						"networkconfig.devices.0.addresses.0", testControlIPs[0]+"/24"),
					resource.TestCheckResourceAttr("talos_control_node.control_0",
						"networkconfig.devices.0.routes.0.network", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("talos_control_node.control_0",
						"networkconfig.devices.0.routes.0.gateway", gateway),
					resource.TestCheckResourceAttr("talos_control_node.control_0",
						"networkconfig.nameservers.0", nameserver),

					testAccTalosConnectivity(testConnArg{
						resourcepath: testControlNodePath(0),
						talosIP:      testControlIPs[0],
					}),
					testAccKubernetesConnectivity("https://"+testControlIPs[0]+":6443"),
					testAccTalosHealth(&clusterNodes{
						Control: []string{testControlIPs[0]},
					}),
					testAccEnsureNMembers(1, testControlIPs[0]),
				),
			},
		},
	})
}

func TestAccResourceTalosControlThreeMaster(t *testing.T) {
	ips := []string{}
	for current := testInitialIPs.From(); current != testInitialIPs.To(); current = current.Next() {
		ips = append(ips, current.String())
	}

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
					Endpoint: testControlIPs[1],
				}) + testControlConfig(&testNode{
					IP:          testControlIPs[0],
					ProvisionIP: ips[0],
					Index:       0,
					Bootstrap:   false,
				}, &testNode{
					IP:          testControlIPs[1],
					ProvisionIP: ips[1],
					Index:       1,
					Bootstrap:   true,
				}, &testNode{
					IP:          testControlIPs[2],
					ProvisionIP: ips[2],
					Index:       2,
					Bootstrap:   false,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccTalosConnectivity(testConnArg{
						resourcepath: testControlNodePath(0),
						talosIP:      testControlIPs[0],
					}, testConnArg{
						resourcepath: testControlNodePath(1),
						talosIP:      testControlIPs[1],
					}, testConnArg{
						resourcepath: testControlNodePath(2),
						talosIP:      testControlIPs[2],
					}),
					testAccKubernetesConnectivity("https://"+testControlIPs[1]+":6443"),
					testAccTalosHealth(&clusterNodes{
						Control: []string{testControlIPs[0], testControlIPs[1], testControlIPs[2]},
					}),
					testAccEnsureNMembers(3, testControlIPs[1]),
				),
			},
		},
	})
}

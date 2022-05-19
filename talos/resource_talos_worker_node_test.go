package talos

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceTalosWorker_basic(t *testing.T) {
	rName := "basic-worker-create-and-update" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rPath := fmt.Sprintf("talos_worker_node.%s", rName)
	ips := genIPsNoCollision("192.168.122.", 2, 126, 2)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckTalosWorkerDestroy,
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				VersionConstraint: "0.6.14",
				Source:            "dmacvicar/libvirt",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTalosWorker_basic(ips, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rPath, "name", rName),
					// test can connect to the node through the talos api
					testAccTalosConnectivity(rPath, rName),
				),
			},
			// test adding the worker node
		},
	})
}

func testAccCheckTalosWorkerDestroy(s *terraform.State) error {
	// TODO: Figure out how to verify a talos node is destroyed
	return nil
}

func testAccTalosWorker_basic(ip []string, rName string) string {
	return talosConfig_basic(ip[0], "https://"+ip[0]+":6443") + controlResource_basic("workerAcc_basic_control", rName, true, ip[0]) + workerResource_basic("workerAcc_basic_worker", rName, ip[1])
}

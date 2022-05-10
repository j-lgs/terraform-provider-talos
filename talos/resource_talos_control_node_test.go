package talos

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// More to test that the basic machinery of the acceptance testing framework works properly
func TestAccResourceTalosControl_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccTalosProviderFactory(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				VersionConstraint: "0.6.14",
				Source:            "dmacvicar/libvirt",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: tes(),
				Check: resource.ComposeTestCheckFunc(
					testAccTalosConnectivity("talos_control_node.single_example", "single_example"),
					testAccKubernetesConnectivity("talos_control_node.single_example", "single_example"),
				),
			},
		},
	})
}

func TestAccResourceTalosControlCreateAndUpdate(t *testing.T) {
	rName := "control-create-update" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rPath := fmt.Sprintf("talos_control_node.%s", rName)
	ips := genIPsNoCollision("192.168.122.", 2, 126, 1)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckTalosControlDestroy,
		Providers:    testAccTalosProviderFactory(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				VersionConstraint: "0.6.14",
				Source:            "dmacvicar/libvirt",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTalosControl_basic(ips[0], rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rPath, "name", rName),
					testAccTalosConnectivity(rPath, rName),
					testAccKubernetesConnectivity(rPath, rName),
				),
			},
			// TODO test update of resource in particular the yaml's change
		},
	})
}

func TestAccResourceTalosControlClusterCreateAndUpdate(t *testing.T) {
	rName := "control-cluster-create-update" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rPath := fmt.Sprintf("talos_control_node.%s", rName)
	ips := genIPsNoCollision("192.168.122.", 2, 126, 3)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckTalosControlDestroy,
		Providers:    testAccTalosProviderFactory(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				VersionConstraint: "0.6.14",
				Source:            "dmacvicar/libvirt",
			},
		},
		Steps: []resource.TestStep{
			{

				Config: testAccTalosControl_cluster(ips, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rPath+"-0", "name", rName+"-0"),
					resource.TestCheckResourceAttr(rPath+"-1", "name", rName+"-1"),
					resource.TestCheckResourceAttr(rPath+"-2", "name", rName+"-2"),
					testAccTalosConnectivity(rPath+"-0", rName+"-0"),
					testAccTalosConnectivity(rPath+"-1", rName+"-1"),
					testAccTalosConnectivity(rPath+"-2", rName+"-2"),
					testAccKubernetesConnectivity(rPath+"-0", rName+"-0"),
					testAccKubernetesConnectivity(rPath+"-1", rName+"-1"),
					testAccKubernetesConnectivity(rPath+"-2", rName+"-2"),
				),
			},
			// TODO test update of resource in particular the yaml's change
		},
	})
}

func TestAccResourceTalosControlStaticPodCreateAndUpdate(t *testing.T) {
	rName := "staticpod-create-update-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rPath := fmt.Sprintf("talos_control_node.%s", rName)
	ips := genIPsNoCollision("192.168.122.", 2, 126, 1)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckTalosControlDestroy,
		Providers:    testAccTalosProviderFactory(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				VersionConstraint: "0.6.14",
				Source:            "dmacvicar/libvirt",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTalosControl_pods(ips[0], rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rPath, "name", rName),
					// Test whether the node has been created successfully
					testAccTalosConnectivity(rPath, rName),
					// Test whether the static pod has been applied successfully and is working.
					testAccHaproxyUp(rPath, rName),
				),
			},
			// TODO test update of resource in particular the yaml's change
		},
	})
}

func testAccPrint(str string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println(str)
		return nil
	}
}

func testAccHaproxyUp(path string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[path]
		if !ok {
			return fmt.Errorf("Not found: %s", path)
		}

		return testHaproxyUp(rs, path, name)
	}
}

func testHaproxyUp(rs *terraform.ResourceState, path string, name string) error {
	is := rs.Primary
	cidr, ok := is.Attributes["interface.0.addresses.0"]
	if !ok {
		return fmt.Errorf("testHaproxyUp: Unable to get interface 0, ip address 0, from resource")
	}

	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("testHaproxyUp: Must provide a valid CIDR IP address, got \"%s\", error \"%s\"", cidr, err.Error())
	}

	host := "http://" + ip.String() + ":8080/haproxy?stats"

	return testHttpUp(host)
}

func testHttpUp(host string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	httpclient := http.Client{}

	var readyresp *http.Response
	for {
		r, err := httpclient.Get(host)
		if err == nil && r.StatusCode >= 200 && r.StatusCode <= 299 {
			readyresp = r
			break
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf(ctx.Err().Error() + " - Reason - " + err.Error())
		default:
			time.Sleep(2 * time.Second)
		}
	}

	defer readyresp.Body.Close()

	return nil
}

func testAccCheckTalosControlDestroy(s *terraform.State) error {
	// TODO: Figure out how to verify a talos node is destroyed
	return nil
}

func testAccTalosControl_basic(ip string, rName string) string {
	return talosConfig_basic(ip, "https://"+ip+":6443") + controlResource_basic("controlAcc_control", rName, true, ip)
}

func testAccTalosControl_cluster(ip []string, rName string) string {
	return talosConfig_basic(ip[0], "https://"+ip[1]+":6443") + controlResource_basic("controlAcc_control", rName+"-0", false, ip[0]) + controlResource_basic("controlAcc_control", rName+"-1", true, ip[1]) + controlResource_basic("controlAcc_control", rName+"-2", false, ip[2])
}

func testAccTalosControl_pods(ip string, rName string) string {
	cfg := talosConfig_basic(ip, "https://"+ip+":6443") + controlResource_basic("controlAcc_control", rName, true, ip, `
  kubelet_extra_mount {
	destination = "/var/static-confs"
	type = "bind"
	source = "/var/static-confs"
	options = [
	  "rbind",
	  "rshared",
	  "rw"
	]
  }

  file {
	content = <<EOT
global
  log         /dev/log local0
  log         /dev/log local1 notice
  daemon
defaults
  mode                    tcp
  log                     global
  option                  tcplog
  option                  tcp-check
  option                  dontlognull
  retries                 3
  timeout client          20s
  timeout server          20s
  timeout check           10s
  timeout queue           20s
  option                  redispatch
  timeout connect         5s
frontend http_stats
  bind `+ip+`:8080
  mode http
  stats uri /haproxy?stats
EOT
	permissions = 438
	path = "/var/static-confs/haproxy/haproxy.cfg"
	op = "create"
  }

pod = [<<EOT
apiVersion: v1
kind: Pod
metadata:
 name: haproxy
 namespace: kube-system
spec:
  containers:
  - image: haproxy:2.5.6
    name: haproxy-controlplane
    volumeMounts:
    - mountPath: /usr/local/etc/haproxy/haproxy.cfg
      name: haproxyconf
      readOnly: true
  hostNetwork: true
  volumes:
  - hostPath:
      path: /var/static-confs/haproxy/haproxy.cfg
      type: File
    name: haproxyconf
status: {}
EOT
  ]
`)
	//fmt.Println(cfg)
	return cfg
}

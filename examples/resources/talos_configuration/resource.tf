resource "talos_configuration" {
  name = "taloscluster"

  target_version = "v1.0"
  talos_endpoints = ["192.168.122.100"]
  kubernetes_endpoint = "https://192.168.122.100:6443"
  kubernetes_version = "1.23.6"

  disks = [
	{
	  device_name = "/dev/vdb"
	  partitions = [
		{
		  mount_point = "/var/mnt/partition"
		  size = "50GiB"
		}
	  ]
	}
  ]

  install = {
	disk = "/dev/vda"
  }

  network = [
	{
	  with_dhcpv6 = {
		"eth0": true,
	  }
	  with_vip = {
		"eth1": "192.168.123.200"
	  }
	  with_mtu = {
		"eth1": 9000
	  }
	  with_wireguard = {
		"wg0": {
		  peer = [
			{
			  allowed_ips = [
				"192.168.124.0/24"
			  ]
			  endpoint = "test.endpoint:54302"
			  public_key = "JbHCJXTOS6wRDjZM1an5YHxGz4QsU7VZKim5EBtpMxk="
			}
		  ]
		}
	  }
	}
  ]

  persist = true
  debug = true
}

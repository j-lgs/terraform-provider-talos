package talos

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func TestExpandProxyConfig(t *testing.T) {
	t.Fatalf("Implement test")
}

func TestExpandApiConfig(t *testing.T) {
	t.Fatalf("Implement test")
}

func TestExpandTypeMap(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected map[string]string
	}{
		{
			input:    map[string]interface{}{},
			expected: map[string]string{},
		},
		{
			input: map[string]interface{}{
				"key":        "value",
				"second_key": "second_value",
			},
			expected: map[string]string{
				"key":        "value",
				"second_key": "second_value",
			},
		},
	}

	for _, c := range cases {
		val := ExpandTypeMap[string](c.input.(TypeMap))
		if val == nil {
			t.Fatalf("Returned nil. Must return an empty data structure")
		}
		if !reflect.DeepEqual(val, c.expected) {
			t.Fatalf("Error matching output and expected: %#v - %#v", val, c.expected)
		}
	}
}

func TestExpandTypeList(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected []string
	}{
		{
			input:    []interface{}{},
			expected: []string{},
		},
		{
			input: []interface{}{
				"val_a",
				"val_b",
			},
			expected: []string{
				"val_a",
				"val_b",
			},
		},
	}

	for _, c := range cases {
		val := ExpandTypeList[string](c.input.(TypeList))
		if val == nil {
			t.Fatalf("Returned nil. Must return an empty data structure")
		}
		if !reflect.DeepEqual(val, c.expected) {
			t.Fatalf("Error matching output and expected: %#v - %#v", val, c.expected)
		}
	}
}

func TestExpandDeviceList(t *testing.T) {
	cases := []struct {
		name     string
		input    []interface{}
		expected []*v1alpha1.Device
	}{
		{
			name: "No wireguard",
			input: []interface{}{
				map[string]interface{}{
					"addresses": []interface{}{
						"192.168.0.122/24",
						"192.168.0.123/24",
					},
					"vip":       []interface{}{},
					"wireguard": []interface{}{},
					"bond":      []interface{}{},
					"vlan":      []interface{}{},
					"route": []interface{}{
						map[string]interface{}{
							"gateway": "192.168.0.1",
							"network": "0.0.0.0/0",
						},
						map[string]interface{}{
							"gateway": "192.168.0.2",
							"network": "0.0.0.0/0",
						},
					},
					"name": "eth0",
				},
				map[string]interface{}{
					"addresses": []interface{}{
						"192.168.0.122/24",
						"192.168.0.123/24",
					},
					"vip":       []interface{}{},
					"wireguard": []interface{}{},
					"bond":      []interface{}{},
					"vlan":      []interface{}{},
					"route": []interface{}{
						map[string]interface{}{
							"gateway": "192.168.0.1",
							"network": "0.0.0.0/0",
						},
						map[string]interface{}{
							"gateway": "192.168.0.2",
							"network": "0.0.0.0/0",
						},
					},
					"name": "eth0",
				},
			},
			expected: []*v1alpha1.Device{
				{
					DeviceInterface: "eth0",
					DeviceAddresses: []string{"192.168.0.122/24", "192.168.0.123/24"},
					DeviceRoutes: []*v1alpha1.Route{
						{
							RouteGateway: "192.168.0.1",
							RouteNetwork: "0.0.0.0/0",
						},
						{
							RouteGateway: "192.168.0.2",
							RouteNetwork: "0.0.0.0/0",
						},
					},
				},
				{
					DeviceInterface: "eth0",
					DeviceAddresses: []string{"192.168.0.122/24", "192.168.0.123/24"},
					DeviceRoutes: []*v1alpha1.Route{
						{
							RouteGateway: "192.168.0.1",
							RouteNetwork: "0.0.0.0/0",
						},
						{
							RouteGateway: "192.168.0.2",
							RouteNetwork: "0.0.0.0/0",
						},
					},
				},
			},
		},
		{
			name: "wireguard",
			input: []interface{}{
				map[string]interface{}{
					"addresses": []interface{}{
						"192.168.0.122/24",
						"192.168.0.123/24",
					},
					"wireguard": []interface{}{},
					"vip":       []interface{}{},
					"bond":      []interface{}{},
					"vlan":      []interface{}{},
					"route": []interface{}{
						map[string]interface{}{
							"gateway": "192.168.0.1",
							"network": "0.0.0.0/0",
						},
						map[string]interface{}{
							"gateway": "192.168.0.2",
							"network": "0.0.0.0/0",
						},
					},
					"name": "eth0",
				},
				map[string]interface{}{
					"addresses": []interface{}{
						"192.168.123.10/24",
						"192.168.123.20/24",
					},
					"vip":  []interface{}{},
					"bond": []interface{}{},
					"vlan": []interface{}{},
					"wireguard": []interface{}{
						map[string]interface{}{
							"peer": []interface{}{
								map[string]interface{}{
									"allowed_ips":                   []interface{}{"192.168.122.0/25"},
									"endpoint":                      "wg_endpoint:44444",
									"persistent_keepalive_interval": 25,
									"public_key":                    "dGVzdGluZyB0ZXN0aW5nCg==",
								},
								map[string]interface{}{
									"allowed_ips":                   []interface{}{"192.168.122.128/25"},
									"endpoint":                      "wg_endpoint_2:4444",
									"persistent_keepalive_interval": 25,
									"public_key":                    "dGVzdGluZyB0ZXN0aW5nIHBlZXIgMgo=",
								},
							},
							"private_key": "c3VwZXIgc3VwZXIgc3VwZXIgc2VjcmV0IGtleQo=",
						},
					},
					"route": []interface{}{},
					"name":  "wg0",
				},
			},
			expected: []*v1alpha1.Device{
				{
					DeviceInterface: "eth0",
					DeviceAddresses: []string{"192.168.0.122/24", "192.168.0.123/24"},
					DeviceRoutes: []*v1alpha1.Route{
						{
							RouteGateway: "192.168.0.1",
							RouteNetwork: "0.0.0.0/0",
						},
						{
							RouteGateway: "192.168.0.2",
							RouteNetwork: "0.0.0.0/0",
						},
					},
				},
				{
					DeviceInterface: "wg0",
					DeviceAddresses: []string{"192.168.123.10/24", "192.168.123.20/24"},
					DeviceWireguardConfig: &v1alpha1.DeviceWireguardConfig{
						WireguardPeers: []*v1alpha1.DeviceWireguardPeer{
							{
								WireguardAllowedIPs: []string{
									"192.168.122.0/25",
								},
								WireguardEndpoint:                    "wg_endpoint:44444",
								WireguardPersistentKeepaliveInterval: time.Second * 25,
								WireguardPublicKey:                   "dGVzdGluZyB0ZXN0aW5nCg==",
							},
							{
								WireguardAllowedIPs: []string{
									"192.168.122.128/25",
								},
								WireguardEndpoint:                    "wg_endpoint_2:4444",
								WireguardPersistentKeepaliveInterval: time.Second * 25,
								WireguardPublicKey:                   "dGVzdGluZyB0ZXN0aW5nIHBlZXIgMgo=",
							},
						},
						WireguardPrivateKey: "c3VwZXIgc3VwZXIgc3VwZXIgc2VjcmV0IGtleQo=",
					},
				},
			},
		},
		{
			name: "All options",
			input: []interface{}{
				map[string]interface{}{
					"addresses": []interface{}{
						"192.168.0.123/24",
					},
					"vip":  []interface{}{},
					"bond": []interface{}{},
					"vlan": []interface{}{},
					"wireguard": []interface{}{
						map[string]interface{}{
							"peer": []interface{}{
								map[string]interface{}{
									"allowed_ips":                   []interface{}{"192.168.122.0/25"},
									"endpoint":                      "wg_endpoint:44444",
									"persistent_keepalive_interval": 25,
									"public_key":                    "dGVzdGluZyB0ZXN0aW5nCg==",
								},
							},
							"private_key": "c3VwZXIgc3VwZXIgc3VwZXIgc2VjcmV0IGtleQo=",
						},
					},
					"route": []interface{}{
						map[string]interface{}{
							"gateway": "192.168.0.2",
							"network": "0.0.0.0/0",
						},
					},
					"name": "eth0",
				},
			},
			expected: []*v1alpha1.Device{
				{
					DeviceInterface: "eth0",
					DeviceAddresses: []string{"192.168.0.123/24"},
					DeviceWireguardConfig: &v1alpha1.DeviceWireguardConfig{
						WireguardPeers: []*v1alpha1.DeviceWireguardPeer{
							{
								WireguardAllowedIPs: []string{
									"192.168.122.0/25",
								},
								WireguardEndpoint:                    "wg_endpoint:44444",
								WireguardPersistentKeepaliveInterval: time.Second * 25,
								WireguardPublicKey:                   "dGVzdGluZyB0ZXN0aW5nCg==",
							},
						},
						WireguardPrivateKey: "c3VwZXIgc3VwZXIgc3VwZXIgc2VjcmV0IGtleQo=",
					},

					DeviceRoutes: []*v1alpha1.Route{
						{
							RouteGateway: "192.168.0.2",
							RouteNetwork: "0.0.0.0/0",
						},
					},
				},
			},
		},
	}
	errorCases := []struct {
		name     string
		input    []interface{}
		expected error
	}{
		{
			name: "Several wireguard",
			input: []interface{}{
				map[string]interface{}{
					"addresses": []interface{}{
						"192.168.123.20/24",
					},
					"vip":  []interface{}{},
					"bond": []interface{}{},
					"vlan": []interface{}{},
					"wireguard": []interface{}{
						map[string]interface{}{
							"peer": []interface{}{
								map[string]interface{}{
									"allowed_ips":                   []interface{}{"192.168.122.128/25"},
									"endpoint":                      "wg_endpoint_2:4444",
									"persistent_keepalive_interval": 25,
									"public_key":                    "dGVzdGluZyB0ZXN0aW5nIHBlZXIgMgo=",
								},
							},
							"private_key": "c3VwZXIgc3VwZXIgc3VwZXIgc2VjcmV0IGtleQo=",
						},
						map[string]interface{}{
							"peer": []interface{}{
								map[string]interface{}{
									"allowed_ips":                   []interface{}{"192.168.122.128/25"},
									"endpoint":                      "wg_endpoint_2:4444",
									"persistent_keepalive_interval": 25,
									"public_key":                    "dGVzdGluZyB0ZXN0aW5nIHBlZXIgMgo=",
								},
							},
							"private_key": "c3VwZXIgc3VwZXIgc3VwZXIgc2VjcmV0IGtleQo=",
						},
					},
					"route": []interface{}{},
					"name":  "wg0",
				},
			},
			expected: WireguardExtraFieldError,
		},
	}
	for _, c := range errorCases {
		_, err := ExpandDeviceList(c.input)
		if !errors.Is(err, c.expected) {
			t.Fatalf("Error in case %s: Unexpected error:\n%s - %s", c.name, c.expected.Error(), err.Error())
		}
	}

	for _, c := range cases {
		val, _ := ExpandDeviceList(c.input)
		if !reflect.DeepEqual(val, c.expected) {
			a, _ := json.MarshalIndent(val, "", "  ")
			b, _ := json.MarshalIndent(c.expected, "", "  ")
			t.Fatalf("Error in case %s: matching output and expected:\n%#v - %#v\nValues:\n%s - %s", c.name, val, c.expected, a, b)
		}
	}
}

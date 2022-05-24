package talos

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CertBundle struct {
	AdminCRT         types.String `tfsdk:"admin_crt"`
	AdminKey         types.String `tfsdk:"admin_key"`
	EtcdCRT          types.String `tfsdk:"etcd_crt"`
	EtcdKey          types.String `tfsdk:"etcd_key"`
	K8sCRT           types.String `tfsdk:"k8s_crt"`
	K8sKey           types.String `tfsdk:"k8s_key"`
	K8sAggregatorCRT types.String `tfsdk:"k8s_aggregator_crt"`
	K8sAggregatorKey types.String `tfsdk:"k8s_aggregator_key"`
	K8sServiceKey    types.String `tfsdk:"k8s_service_key"`
	OSCRT            types.String `tfsdk:"os_crt"`
	OSKey            types.String `tfsdk:"os_key"`
}

var CertBundleSchema = tfsdk.Schema{
	MarkdownDescription: "Represents the keys and certificates throughout Talos.",
	Attributes: map[string]tfsdk.Attribute{
		"admin_crt": {
			Optional:            true,
			Type:                types.StringType,
			MarkdownDescription: "PEM encoded cluster admin crt.",
		},
		"admin_key": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "PEM encoded cluster admin key.",
		},
		"etcd_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded etcd crt.",
		},
		"etcd_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded etcd key.",
		},
		"k8s_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded crt for k8s..",
		},
		"k8s_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for k8s.",
		},
		"k8s_aggregator_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded crt for the k8s aggregator.",
		},
		"k8s_aggregator_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for the k8s aggregator.",
		},
		"k8s_service_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for the k8s service.",
		},
		"os_crt": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded crt for OS.",
		},
		"os_key": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "PEM encoded key for OS.",
		},
	},
}

type NetworkConfigOptions struct {
	Kubespan      types.Bool              `tfsdk:"with_kubespan"`
	VIP           map[string]types.String `tfsdk:"with_vip"`
	Wireguard     map[string]Wireguard    `tfsdk:"with_wireguard"`
	MTU           map[string]types.Int64  `tfsdk:"with_mtu"`
	CIDR          map[string]types.String `tfsdk:"with_cidr"`
	DHCPv6        map[string]types.Bool   `tfsdk:"with_dhcpv6"`
	DHCPv4        map[string]types.Bool   `tfsdk:"with_dhcpv4"`
	DHCP          map[string]types.Bool   `tfsdk:"with_dhcp"`
	Ignore        map[string]types.Bool   `tfsdk:"with_ignore"`
	Nameservers   []types.String          `tfsdk:"with_nameservers"`
	NetworkConfig *NetworkConfig          `tfsdk:"with_networkconfig"`
}

var NetworkConfigOptionSchema = tfsdk.Schema{
	MarkdownDescription: "Represents globally applied network configuration options.",
	Attributes: map[string]tfsdk.Attribute{
		"with_kubespan": {
			Type:     types.BoolType,
			Optional: true,
		},
		"with_vip": {
			Optional:    true,
			Description: "Configures an interface for Virtual shared IP. Maps interfaces names to desired CIDRs.",
			Type: types.MapType{
				ElemType: types.StringType,
			},
		},
		"with_wireguard": {
			Optional:    true,
			Description: WireguardSchema.Description,
			Attributes:  tfsdk.MapNestedAttributes(WireguardSchema.Attributes, tfsdk.MapNestedAttributesOptions{}),
		},
		"with_mtu": {
			Optional:    true,
			Description: "Configures an interface's MTU.",
			Type: types.MapType{
				ElemType: types.Int64Type,
			},
		},
		"with_cidr": {
			Optional:    true,
			Description: "Configures an interface for static addressing.",
			Type: types.MapType{
				ElemType: types.StringType,
			},
		},
		"with_dhcpv6": {
			Optional:    true,
			Description: "Enables DHCPv6 for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_dhcpv4": {
			Optional:    true,
			Description: "Enables DHCPv4 for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_dhcp": {
			Optional:    true,
			Description: "Enables DHCP for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_ignore": {
			Optional:    true,
			Description: "Enables DHCP for an interface.",
			Type: types.MapType{
				ElemType: types.BoolType,
			},
		},
		"with_nameservers": {
			Optional:    true,
			Description: "Sets global nameservers list.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
		},
		"with_networkconfig": {
			Optional:    true,
			Description: NetworkConfigSchema.Description,
			Attributes:  tfsdk.SingleNestedAttributes(NetworkConfigSchema.Attributes),
		},
	},
}

type SecretBundle struct {
	ID             types.String `tfsdk:"id"`
	CertBundle     *CertBundle  `tfsdk:"cert_bundle"`
	Secret         types.String `tfsdk:"secret"`
	BootstrapToken types.String `tfsdk:"bootstrap_token"`
	AESEncryption  types.String `tfsdk:"aes_cbc_encryption"`
	TrustdToken    types.String `tfsdk:"trustd_token"`
}

var SecretBundleSchema = tfsdk.Schema{
	MarkdownDescription: "Represents secrets used throughout a Talos install.",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Unique cluster ID for Talos. Base64 encoded binary data.",
		},
		"cert_bundle": {
			Optional:            true,
			MarkdownDescription: CertBundleSchema.MarkdownDescription,
			Attributes:          tfsdk.SingleNestedAttributes(CertBundleSchema.Attributes),
		},
		"secret": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique cluster secret for Talos. Base64 encoded binary data.",
		},
		"bootstrap_token": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique token for Talos bootstrap.",
		},
		"aes_cbc_encryption": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique secret for Talos disk encryption. Base64 encoded binary data.",
		},
		"trustd_token": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Unique token for Talos trustd.",
		},
	},
}

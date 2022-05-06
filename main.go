package main

import (
	"terraform-provider-talos/talos"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return talos.Provider()
		},
	}

	plugin.Serve(opts)
}

package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-talos/talos"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		Address: "registry.terraform.io/hashicorp/scaffolding",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), talos.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}

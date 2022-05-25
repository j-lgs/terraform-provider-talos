package datatypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

// Data copies data from terraform state types to talos types.
func (planManifest InlineManifest) Data() (interface{}, error) {
	manifest := v1alpha1.ClusterInlineManifest{}

	if planManifest.Name.Value != "" {
		manifest.InlineManifestName = planManifest.Name.Value
		manifest.InlineManifestContents = planManifest.Content.Value
	}

	return manifest, nil
}

// Read copies data from talos types to terraform state types.
func (planManifest *InlineManifest) Read(talosInlineManifest interface{}) error {
	manifest := talosInlineManifest.(v1alpha1.ClusterInlineManifest)
	if manifest.InlineManifestName != "" {
		planManifest.Name = types.String{Value: manifest.InlineManifestName}
		planManifest.Content = types.String{Value: manifest.InlineManifestContents}
	}

	return nil
}

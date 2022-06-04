package datatypes

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func TestKubespanData(t *testing.T) {
	tests := []struct {
		input  NetworkKubeSpan
		output v1alpha1.NetworkKubeSpan
	}{
		{
			input: NetworkKubeSpan{
				Enabled: types.Bool{Value: true},
			},
			output: v1alpha1.NetworkKubeSpan{
				KubeSpanEnabled: true,
			},
		}, {
			input:  NetworkKubeSpanExample,
			output: TalosNetworkKubeSpanExample,
		},
	}

	for _, tc := range tests {
		res, err := tc.input.Data()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(res.(v1alpha1.NetworkKubeSpan), tc.output) {
			t.Fatalf("expected: %v, got: %v", tc.output, res)
		}
	}
}

func TestKubespanRead(t *testing.T) {
	tests := []struct {
		input  v1alpha1.NetworkKubeSpan
		output NetworkKubeSpan
	}{
		{
			input: v1alpha1.NetworkKubeSpan{
				KubeSpanEnabled: true,
			},
			output: NetworkKubeSpan{
				Enabled: types.Bool{Value: true},
			},
		}, {
			input:  TalosNetworkKubeSpanExample,
			output: NetworkKubeSpanExample,
		},
	}

	for _, tc := range tests {
		kubespan := &NetworkKubeSpan{}
		err := kubespan.Read(&tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(*kubespan, tc.output) {
			t.Fatalf("expected: %v, got: %v", tc.output, *kubespan)
		}
	}
}

package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/talos-systems/talos/pkg/machinery/api/machine"
	"github.com/talos-systems/talos/pkg/machinery/api/resource"
	v1alpha1 "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	machinetype "github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type nodeResourceData interface {
	TalosData(*v1alpha1.Config) (*v1alpha1.Config, error)
	ReadInto(*v1alpha1.Config) error
	Generate() error
}

type readData struct {
	ConfigIP   string
	BaseConfig string
}

func readConfig[N nodeResourceData](ctx context.Context, nodeData N, data readData) (out *v1alpha1.Config, errDesc string, err error) {
	host := net.JoinHostPort(data.ConfigIP, strconv.Itoa(talosPort))

	input := generate.Input{}
	if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
		return nil, "Unable to marshal node's base_config data into it's generate.Input struct.", err
	}

	conn, err := secureConn(ctx, input, host)
	if err != nil {
		return nil, "Unable to make a secure connection to read the node's Talos config.", err
	}

	defer conn.Close()
	client := resource.NewResourceServiceClient(conn)
	resourceResp, err := client.Get(ctx, &resource.GetRequest{
		Type:      "MachineConfig",
		Namespace: "config",
		Id:        "v1alpha1",
	})
	if err != nil {
		return nil, "Error getting Machine Configuration", err
	}

	if len(resourceResp.Messages) < 1 {
		return nil, "Invalid message count.",
			fmt.Errorf("invalid message count from the Talos resource get request. Expected > 1 but got %d", len(resourceResp.Messages))
	}

	out = &v1alpha1.Config{}
	err = yaml.Unmarshal(resourceResp.Messages[0].Resource.Spec.Yaml, out)
	if err != nil {
		return nil, "Unable to unmarshal Talos configuration into it's struct.", err
	}

	return
}

type configData struct {
	Bootstrap   bool
	CreateNode  bool
	Mode        machine.ApplyConfigurationRequest_Mode
	ProvisionIP string
	ConfigIP    string
	BaseConfig  string
	MachineType machinetype.Type
}

func genConfig[N nodeResourceData](machineType machinetype.Type, input *generate.Input, nodeData *N) (out string, err error) {
	cfg, err := generate.Config(machineType, input)
	if err != nil {
		err = fmt.Errorf("failed to generate Talos configuration struct for node: %w", err)
		return
	}

	(*nodeData).Generate()
	newCfg, err := (*nodeData).TalosData(cfg)
	if err != nil {
		err = fmt.Errorf("failed to generate configuration: %w", err)
		return
	}

	var confYaml []byte
	confYaml, err = newCfg.Bytes()
	if err != nil {
		err = fmt.Errorf("failed to generate config yaml: %w", err)
		return
	}

	out = string(regexp.MustCompile(`\s*#.*`).ReplaceAll(confYaml, nil))
	return
}

func applyConfig[N nodeResourceData](ctx context.Context, nodeData *N, data configData) (out string, errDesc string, err error) {
	input := generate.Input{}
	if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
		return "", "Failed to unmarshal input bundle", err
	}

	yaml, err := genConfig(data.MachineType, &input, nodeData)
	if err != nil {
		return "", "error rendering configuration YAML", err
	}

	var conn *grpc.ClientConn
	if data.CreateNode {
		ip := data.ProvisionIP
		host := net.JoinHostPort(ip, strconv.Itoa(talosPort))
		conn, err = insecureConn(ctx, host)
		if err != nil {
			return "", "Unable to make insecure connection to Talos machine. Ensure it is in maintainence mode.", err
		}
	} else {
		ip := data.ConfigIP
		host := net.JoinHostPort(ip, strconv.Itoa(talosPort))
		input := generate.Input{}
		if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
			return "", "Unable to unmarshal BaseConfig json into a Talos Input struct.", err
		}

		conn, err = secureConn(ctx, input, host)
		if err != nil {
			return "", "Unable to make secure connection to the Talos machine.", err
		}
	}

	defer conn.Close()
	client := machine.NewMachineServiceClient(conn)
	_, err = client.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: []byte(yaml),
		Mode: machine.ApplyConfigurationRequest_Mode(data.Mode),
	})
	if err != nil {
		return "", "Error applying configuration", err
	}

	if data.MachineType == machinetype.TypeControlPlane && data.Bootstrap {
		// Wait for time to be synchronised.
		time.Sleep(5 * time.Second)

		ip := data.ConfigIP
		host := net.JoinHostPort(ip, strconv.Itoa(talosPort))
		input := generate.Input{}
		if err := json.Unmarshal([]byte(data.BaseConfig), &input); err != nil {
			return "", "Unable to unmarshal BaseConfig json into a Talos Input struct.", err
		}

		conn, err := secureConn(ctx, input, host)
		if err != nil {
			return "", "Unable to make secure connection to the Talos machine.", err
		}
		defer conn.Close()
		client := machine.NewMachineServiceClient(conn)
		_, err = client.Bootstrap(ctx, &machine.BootstrapRequest{})
		if err != nil {
			return "", "Error attempting to bootstrap the machine.", err
		}
	}

	return
}

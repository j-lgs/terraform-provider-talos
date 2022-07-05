package talos

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
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

func genConfig[N nodeResourceData](machineType machinetype.Type, input *generate.Input, nodeData N) ([]byte, error) {
	cfg, err := generate.Config(machineType, input)
	if err != nil {
		return nil, err
	}

	newCfg, err := nodeData.TalosData(cfg)
	if err != nil {
		return nil, err
	}

	confYaml, err := newCfg.Bytes()
	if err != nil {
		return nil, err
	}

	// strip all comments from the generated yaml
	rexp, err := regexp.Compile(`\s*#.*`)
	if err != nil {
		return nil, err
	}

	out := rexp.ReplaceAll(confYaml, nil)

	return out, nil
}

func applyConfig(ctx context.Context, conn *grpc.ClientConn, yaml []byte, mode machine.ApplyConfigurationRequest_Mode) error {
	defer conn.Close()

	client := machine.NewMachineServiceClient(conn)
	_, err := client.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: yaml,
		Mode: mode,
	})
	if err != nil {
		return err
	}

	return nil
}

func bootstrap(ctx context.Context, conn *grpc.ClientConn) error {
	defer conn.Close()

	// Wait for time to be synchronised after installation.
	// TODO: Figure out a better way of handling this. most likely by polling
	// api endpoint.
	time.Sleep(15 * time.Second)
	// Require more time if inside a Github Action
	if _, set := os.LookupEnv("GITHUB_ACTIONS"); set {
		time.Sleep(60 * time.Second)
	}

	client := machine.NewMachineServiceClient(conn)
	_, err := client.Bootstrap(ctx, &machine.BootstrapRequest{})
	if err != nil {
		return err
	}

	return nil
}

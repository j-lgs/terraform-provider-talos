package talos

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func checkArp(mac string) (net.IP, error) {
	arp, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, err
	}
	defer arp.Close()

	scanner := bufio.NewScanner(arp)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if strings.EqualFold(f[3], mac) {
			return net.ParseIP(f[0]), nil
		}
	}

	return nil, nil
}

func lookupIP(ctx context.Context, network string, mac string) (net.IP, error) {
	// Check if it's in the initial table

	var ip net.IP
	var err error

	if ip, err = checkArp(mac); err != nil {
		return nil, err
	}
	if ip != nil {
		return ip, err
	}

	ctx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	for poll := true; poll; poll = (ip == nil) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			err := exec.CommandContext(ctx, "nmap", "-sP", network).Run()
			if err != nil {
				return nil, fmt.Errorf("%s", err)
			}
			if ip, err = checkArp(mac); err != nil {
				return nil, err
			}
			if ip != nil {
				return ip, nil
			}
			time.Sleep(5 * time.Second)
		}
	}

	return ip, nil
}

func makeTLSConfig(certs generate.Certs, secure bool) (tls.Config, error) {
	tlsConfig := &tls.Config{}
	if secure {
		clientCert, err := tls.X509KeyPair(certs.Admin.Crt, certs.Admin.Key)
		if err != nil {
			return tls.Config{}, err
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(certs.OS.Crt); !ok {
			return tls.Config{}, fmt.Errorf("unable to append certs from PEM")
		}

		return tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{clientCert},
		}, nil
	}

	tlsConfig.InsecureSkipVerify = true
	return tls.Config{
		InsecureSkipVerify: true,
	}, nil

}

func waitTillTalosMachineUp(ctx context.Context, tlsConfig *tls.Config, host string, secure bool) error {
	tflog.Info(ctx, "Waiting for talos machine to be up")
	// overall timeout should be 5 mins

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithBlock(),
	}
	ctx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	for _, err := grpc.Dial(host, opts...); err != nil; {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			tflog.Info(ctx, "Retrying connection to "+host+" reason "+err.Error())
			time.Sleep(5 * time.Second)
		}
	}

	tflog.Info(ctx, "Waiting for talos machine to be up")

	return nil
}

func insecureConn(ctx context.Context, host string) (*grpc.ClientConn, error) {
	tlsConfig, err := makeTLSConfig(generate.Certs{}, false)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)),
	}

	waitTillTalosMachineUp(ctx, &tlsConfig, host, false)

	conn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		tflog.Error(ctx, "Error dailing talos.")
		return nil, err
	}

	return conn, nil
}

func secureConn(ctx context.Context, input generate.Input, host string) (*grpc.ClientConn, error) {
	tlsConfig, err := makeTLSConfig(*input.Certs, true)
	if err != err {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)),
	}

	waitTillTalosMachineUp(ctx, &tlsConfig, host, true)

	conn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		tflog.Error(ctx, "Error securely dailing talos.")
		return nil, err
	}

	return conn, nil
}

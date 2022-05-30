package talos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

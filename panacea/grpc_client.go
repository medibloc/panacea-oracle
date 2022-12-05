package panacea

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/cosmos/cosmos-sdk/types/tx"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCClient struct {
	conn *grpc.ClientConn
}

func NewGRPCClient(grpcAddr string) (*GRPCClient, error) {
	log.Infof("dialing to Panacea gRPC endpoint: %s", grpcAddr)

	parsedUrl, err := url.Parse(grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gRPC endpoint. please use absolute URL (scheme://host:port): %w", err)
	}

	var cred grpc.DialOption

	if parsedUrl.Scheme == "https" {
		cred = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	} else {
		cred = grpc.WithInsecure()
	}

	conn, err := grpc.Dial(parsedUrl.Host, cred)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Panacea: %w", err)
	}

	return &GRPCClient{
		conn: conn,
	}, nil
}

func (c *GRPCClient) Close() error {
	log.Info("closing Panacea gRPC connection")
	return c.conn.Close()
}

func (c *GRPCClient) BroadcastTx(txBytes []byte) (*tx.BroadcastTxResponse, error) {
	txClient := tx.NewServiceClient(c.conn)

	return txClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		},
	)
}

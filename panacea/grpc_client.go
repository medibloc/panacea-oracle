package panacea

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCClient struct {
	conn    *grpc.ClientConn
	cdc     *codec.ProtoCodec
	chainID string
}

func NewGRPCClient(grpcAddr, chainID string) (*GRPCClient, error) {
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
		conn:    conn,
		cdc:     codec.NewProtoCodec(makeInterfaceRegistry()),
		chainID: chainID,
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

func (c *GRPCClient) GetCdc() *codec.ProtoCodec {
	return c.cdc
}

func (c *GRPCClient) GetChainID() string {
	return c.chainID
}

func (c *GRPCClient) GetAccount(address string) (authtypes.AccountI, error) {
	client := authtypes.NewQueryClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, fmt.Errorf("failed to get account info via grpc: %w", err)
	}

	var acc authtypes.AccountI
	if err := c.cdc.UnpackAny(response.GetAccount(), &acc); err != nil {
		return nil, fmt.Errorf("failed to unpack account info: %w", err)
	}
	return acc, nil
}

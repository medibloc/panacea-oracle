package service

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/tendermint/tendermint/libs/os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/ipfs"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GRPCClient() *panacea.GRPCClient
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	Config() *config.Config
	QueryClient() panacea.QueryClient
	IPFS() *ipfs.IPFS
	BroadcastTx(...sdk.Msg) (int64, string, error)
	StartSubscriptions(...event.Event) error
	Close() error
}

type service struct {
	conf        *config.Config
	enclaveInfo *sgx.EnclaveInfo

	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	queryClient panacea.QueryClient
	grpcClient  *panacea.GRPCClient
	subscriber  *event.PanaceaSubscriber
	txBuilder   *panacea.TxBuilder
	ipfs        *ipfs.IPFS
}

func New(conf *config.Config) (Service, error) {
	queryClient, err := panacea.LoadVerifiedQueryClient(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to load query client: %w", err)
	}

	return NewWithQueryClient(conf, queryClient)
}

func NewWithQueryClient(conf *config.Config, queryClient panacea.QueryClient) (Service, error) {
	oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
	if err != nil {
		return nil, err
	}

	var oraclePrivKey *btcec.PrivateKey
	if os.FileExists(conf.AbsOraclePrivKeyPath()) {
		oraclePrivKeyBz, err := sgx.UnsealFromFile(conf.AbsOraclePrivKeyPath())
		if err != nil {
			return nil, fmt.Errorf("failed to unseal oracle_priv_key.sealed file: %w", err)
		}
		oraclePrivKey, _ = crypto.PrivKeyFromBytes(oraclePrivKeyBz)
	}

	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to set self-enclave info: %w", err)
	}

	newIpfs, err := ipfs.NewIPFS(conf.IPFS.IPFSNodeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection to IPFS node(%s): %w", conf.IPFS.IPFSNodeAddr, err)
	}

	grpcClient, err := panacea.NewGRPCClient(conf.Panacea.GRPCAddr, conf.Panacea.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gRPC client: %w", err)
	}

	txBuilder := panacea.NewTxBuilder(*grpcClient)

	subscriber, err := event.NewSubscriber(conf.Panacea.RPCAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to init subscriber: %w", err)
	}

	return &service{
		conf:          conf,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		enclaveInfo:   selfEnclaveInfo,
		queryClient:   queryClient,
		grpcClient:    grpcClient,
		txBuilder:     txBuilder,
		subscriber:    subscriber,
		ipfs:          newIpfs,
	}, nil
}

func (s *service) StartSubscriptions(events ...event.Event) error {
	return s.subscriber.Run(events...)
}

func (s *service) Close() error {
	log.Info("calling the service's close function")
	if err := s.queryClient.Close(); err != nil {
		log.Warn(err)
	}
	if err := s.grpcClient.Close(); err != nil {
		log.Warn(err)
	}
	if err := s.subscriber.Close(); err != nil {
		log.Warn(err)
	}

	return nil
}

func (s *service) Config() *config.Config {
	return s.conf
}

func (s *service) OracleAcc() *panacea.OracleAccount {
	return s.oracleAccount
}

func (s *service) OraclePrivKey() *btcec.PrivateKey {
	return s.oraclePrivKey
}

func (s *service) EnclaveInfo() *sgx.EnclaveInfo {
	return s.enclaveInfo
}

func (s *service) GRPCClient() *panacea.GRPCClient {
	return s.grpcClient
}

func (s *service) QueryClient() panacea.QueryClient {
	return s.queryClient
}

func (s *service) BroadcastTx(msg ...sdk.Msg) (int64, string, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(s.Config().Panacea.DefaultFeeAmount)

	txBytes, err := s.txBuilder.GenerateSignedTxBytes(
		s.OracleAcc().GetPrivKey(),
		s.Config().Panacea.DefaultGasLimit,
		defaultFeeAmount,
		msg...,
	)
	if err != nil {
		return 0, "", fmt.Errorf("failed to generate signed Tx bytes: %w", err)
	}

	resp, err := s.GRPCClient().BroadcastTx(txBytes)
	if err != nil {
		return 0, "", fmt.Errorf("broadcast transaction failed. txBytes(%v)", txBytes)
	}

	if resp.TxResponse.Code != 0 {
		return 0, "", fmt.Errorf("transaction failed: %v", resp.TxResponse.RawLog)
	}

	return resp.TxResponse.Height, resp.TxResponse.TxHash, nil
}

func (s *service) IPFS() *ipfs.IPFS {
	return s.ipfs
}

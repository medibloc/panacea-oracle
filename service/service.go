package service

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-oracle/consumer_service"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/tendermint/tendermint/libs/os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GRPCClient() panacea.GRPCClient
	EnclaveInfo() *sgx.EnclaveInfo
	SGX() sgx.Sgx
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	Config() *config.Config
	QueryClient() panacea.QueryClient
	ConsumerService() consumer_service.FileStorage
	BroadcastTx(...sdk.Msg) (int64, string, error)
	StartSubscriptions(...event.Event) error
	Close() error
}

type service struct {
	conf        *config.Config
	enclaveInfo *sgx.EnclaveInfo
	sgx         sgx.Sgx

	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	queryClient     panacea.QueryClient
	grpcClient      panacea.GRPCClient
	consumerService consumer_service.FileStorage
	subscriber      *event.PanaceaSubscriber
	txBuilder       *panacea.TxBuilder
}

func New(conf *config.Config, sgx sgx.Sgx, queryClient panacea.QueryClient) (Service, error) {
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

	selfEnclaveInfo, err := sgx.GenerateSelfEnclaveInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to set self-enclave info: %w", err)
	}

	grpcClient, err := panacea.NewGRPCClient(conf.Panacea.GRPCAddr, conf.Panacea.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gRPC client: %w", err)
	}

	txBuilder := panacea.NewTxBuilder(grpcClient)

	subscriber, err := event.NewSubscriber(conf.Panacea.RPCAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to init subscriber: %w", err)
	}

	consumerService := consumer_service.NewConsumerService(
		oraclePrivKey,
		oracleAccount,
		conf.Consumer.Timeout,
	)

	return &service{
		conf:            conf,
		oracleAccount:   oracleAccount,
		oraclePrivKey:   oraclePrivKey,
		enclaveInfo:     selfEnclaveInfo,
		sgx:             sgx,
		queryClient:     queryClient,
		grpcClient:      grpcClient,
		consumerService: consumerService,
		txBuilder:       txBuilder,
		subscriber:      subscriber,
	}, nil
}

func (s *service) StartSubscriptions(events ...event.Event) error {
	return s.subscriber.Run(events...)
}

func (s *service) Close() error {
	log.Info("calling the service's close function")
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

func (s *service) SGX() sgx.Sgx {
	return s.sgx
}

func (s *service) GRPCClient() panacea.GRPCClient {
	return s.grpcClient
}

func (s *service) QueryClient() panacea.QueryClient {
	return s.queryClient
}

func (s *service) ConsumerService() consumer_service.FileStorage {
	return s.consumerService
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

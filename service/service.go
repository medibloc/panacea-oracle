package service

import (
	"context"
	"fmt"

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

type Service struct {
	conf        *config.Config
	enclaveInfo *sgx.EnclaveInfo

	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	queryClient panacea.QueryClient
	grpcClient  *panacea.GRPCClient
	subscriber  *event.PanaceaSubscriber
	ipfs        *ipfs.IPFS
}

func New(conf *config.Config) (*Service, error) {
	oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
	if err != nil {
		return nil, err
	}

	var oraclePrivKey *btcec.PrivateKey
	if os.FileExists(conf.AbsOraclePubKeyPath()) {
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

	queryClient, err := panacea.LoadVerifiedQueryClient(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to load query client: %w", err)
	}

	grpcClient, err := panacea.NewGRPCClient(conf.Panacea.GRPCAddr)
	if err != nil {
		if err := queryClient.Close(); err != nil {
			log.Warn(err)
		}
		return nil, fmt.Errorf("failed to create a new gRPC client: %w", err)
	}

	subscriber, err := event.NewSubscriber(conf.Panacea.RPCAddr)
	if err != nil {
		if err := queryClient.Close(); err != nil {
			log.Warn(err)
		}
		if err := grpcClient.Close(); err != nil {
			log.Warn(err)
		}
		return nil, fmt.Errorf("failed to init subscriber: %w", err)
	}

	newIpfs := ipfs.NewIPFS(conf.IPFS.IPFSNodeAddr)

	return &Service{
		conf:          conf,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		enclaveInfo:   selfEnclaveInfo,
		queryClient:   queryClient,
		grpcClient:    grpcClient,
		subscriber:    subscriber,
		ipfs:          newIpfs,
	}, nil
}

func (s *Service) StartSubscriptions(events ...event.Event) error {
	return s.subscriber.Run(events...)
}

func (s *Service) Close() error {
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

func (s *Service) Config() *config.Config {
	return s.conf
}

func (s *Service) OracleAcc() *panacea.OracleAccount {
	return s.oracleAccount
}

func (s *Service) OraclePrivKey() *btcec.PrivateKey {
	return s.oraclePrivKey
}

func (s *Service) EnclaveInfo() *sgx.EnclaveInfo {
	return s.enclaveInfo
}

func (s *Service) GRPCClient() *panacea.GRPCClient {
	return s.grpcClient
}

func (s *Service) QueryClient() panacea.QueryClient {
	return s.queryClient
}

func (s *Service) IPFS() *ipfs.IPFS {
	return s.ipfs
}

func (s *Service) BroadcastTx(txBytes []byte) (int64, string, error) {
	resp, err := s.GRPCClient().BroadcastTx(txBytes)
	if err != nil {
		return 0, "", fmt.Errorf("broadcast transaction failed. txBytes(%v)", txBytes)
	}

	if resp.TxResponse.Code != 0 {
		return 0, "", fmt.Errorf("transaction failed: %v", resp.TxResponse.RawLog)
	}

	return resp.TxResponse.Height, resp.TxResponse.TxHash, nil
}

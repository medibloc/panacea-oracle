package oracle

import (
	"fmt"

	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/tendermint/tendermint/libs/os"
)

type oracleService struct {
	*service.Service
}

func NewWithQueryClient(conf *config.Config, queryClient panacea.QueryClient) (*oracleService, error) {
	svc, err := service.NewWithQueryClient(conf, queryClient)
	if err != nil {
		return nil, err
	}

	return &oracleService{svc}, nil
}

func (s *oracleService) GetAndStoreOraclePrivKey() error {
	oraclePrivKeyBz, err := s.getOraclePrivKey()
	if err != nil {
		return err
	}

	if err := sgx.SealToFile(oraclePrivKeyBz, s.Config().AbsOraclePrivKeyPath()); err != nil {
		return fmt.Errorf("failed to seal oraclePrivKey to file. %w", err)
	}

	return nil
}

func (s *oracleService) getOraclePrivKey() ([]byte, error) {
	oraclePrivKeyPath := s.Config().AbsOraclePrivKeyPath()
	if os.FileExists(oraclePrivKeyPath) {
		return nil, fmt.Errorf("the oracle private key already exists")
	}

	shareKeyBz, err := s.deriveSharedKey()
	if err != nil {
		return nil, err
	}

	encryptedOraclePrivKeyBz, err := s.getEncryptedOraclePrivKey()
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(shareKeyBz, nil, encryptedOraclePrivKeyBz)
}

func (s *oracleService) deriveSharedKey() ([]byte, error) {
	nodePrivKeyPath := s.Config().AbsNodePrivKeyPath()
	if !os.FileExists(nodePrivKeyPath) {
		return nil, fmt.Errorf("the node private key is not exists")
	}
	nodePrivKeyBz, err := sgx.UnsealFromFile(nodePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal nodePrivKey from file.%w", err)
	}
	nodePrivKey, _ := crypto.PrivKeyFromBytes(nodePrivKeyBz)

	oraclePublicKey, err := s.QueryClient().GetOracleParamsPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get oraclePublicKey. %w", err)
	}

	shareKeyBz := crypto.DeriveSharedKey(nodePrivKey, oraclePublicKey, crypto.KDFSHA256)
	return shareKeyBz, nil
}

func (s *oracleService) getEncryptedOraclePrivKey() ([]byte, error) {
	uniqueID := s.EnclaveInfo().UniqueIDHex()
	oracleAddress := s.OracleAcc().GetAddress()
	oracleRegistration, err := s.QueryClient().GetOracleRegistration(uniqueID, oracleAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracleRegistration. %w", err)
	}

	return oracleRegistration.EncryptedOraclePrivKey, nil
}

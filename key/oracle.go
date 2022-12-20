package key

import (
	"context"
	"fmt"

	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/tendermint/tendermint/libs/os"
)

func GetAndStoreOraclePrivKey(ctx context.Context, svc service.Service) error {
	oraclePrivKeyBz, err := getOraclePrivKey(ctx, svc)
	if err != nil {
		return err
	}

	if err := sgx.SealToFile(oraclePrivKeyBz, svc.Config().AbsOraclePrivKeyPath()); err != nil {
		return fmt.Errorf("failed to seal oraclePrivKey to file. %w", err)
	}

	return nil
}

func getOraclePrivKey(ctx context.Context, svc service.Service) ([]byte, error) {
	oraclePrivKeyPath := svc.Config().AbsOraclePrivKeyPath()
	if os.FileExists(oraclePrivKeyPath) {
		return nil, fmt.Errorf("the oracle private key already exists")
	}

	shareKeyBz, err := deriveSharedKey(ctx, svc)
	if err != nil {
		return nil, err
	}

	encryptedOraclePrivKeyBz, err := getEncryptedOraclePrivKey(ctx, svc)
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(shareKeyBz, nil, encryptedOraclePrivKeyBz)
}

func deriveSharedKey(ctx context.Context, svc service.Service) ([]byte, error) {
	nodePrivKeyPath := svc.Config().AbsNodePrivKeyPath()
	if !os.FileExists(nodePrivKeyPath) {
		return nil, fmt.Errorf("the node private key is not exists")
	}
	nodePrivKeyBz, err := sgx.UnsealFromFile(nodePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal nodePrivKey from file.%w", err)
	}
	nodePrivKey, _ := crypto.PrivKeyFromBytes(nodePrivKeyBz)

	oraclePublicKey, err := svc.QueryClient().GetOracleParamsPublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get oraclePublicKey. %w", err)
	}

	shareKeyBz := crypto.DeriveSharedKey(nodePrivKey, oraclePublicKey, crypto.KDFSHA256)
	return shareKeyBz, nil
}

func getEncryptedOraclePrivKey(ctx context.Context, svc service.Service) ([]byte, error) {
	uniqueID := svc.EnclaveInfo().UniqueIDHex()
	oracleAddress := svc.OracleAcc().GetAddress()
	oracleRegistration, err := svc.QueryClient().GetOracleRegistration(ctx, uniqueID, oracleAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracleRegistration. %w", err)
	}

	return oracleRegistration.EncryptedOraclePrivKey, nil
}

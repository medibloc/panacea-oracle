package key

import (
	"context"
	"fmt"

	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/service"
	"github.com/tendermint/tendermint/libs/os"
)

func RetrieveAndStoreOraclePrivKey(ctx context.Context, svc service.Service, encryptedOraclePrivKey []byte) error {
	oraclePrivKeyBz, err := retrieveOraclePrivKey(ctx, svc, encryptedOraclePrivKey)
	if err != nil {
		return err
	}

	if err := svc.GetSgx().SealToFile(oraclePrivKeyBz, svc.GetConfig().AbsOraclePrivKeyPath()); err != nil {
		return fmt.Errorf("failed to seal oraclePrivKey to file. %w", err)
	}

	return nil
}

func retrieveOraclePrivKey(ctx context.Context, svc service.Service, encryptedOraclePrivKey []byte) ([]byte, error) {
	oraclePrivKeyPath := svc.GetConfig().AbsOraclePrivKeyPath()
	if os.FileExists(oraclePrivKeyPath) {
		return nil, fmt.Errorf("the oracle private key already exists")
	}

	shareKeyBz, err := deriveSharedKey(ctx, svc)
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(shareKeyBz, nil, encryptedOraclePrivKey)
}

func deriveSharedKey(ctx context.Context, svc service.Service) ([]byte, error) {
	nodePrivKeyPath := svc.GetConfig().AbsNodePrivKeyPath()
	if !os.FileExists(nodePrivKeyPath) {
		return nil, fmt.Errorf("the node private key is not exists")
	}
	nodePrivKeyBz, err := svc.GetSgx().UnsealFromFile(nodePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal nodePrivKey from file.%w", err)
	}
	nodePrivKey, _ := crypto.PrivKeyFromBytes(nodePrivKeyBz)

	oraclePublicKey, err := svc.GetQueryClient().GetOracleParamsPublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get oraclePublicKey. %w", err)
	}

	shareKeyBz := crypto.DeriveSharedKey(nodePrivKey, oraclePublicKey, crypto.KDFSHA256)
	return shareKeyBz, nil
}

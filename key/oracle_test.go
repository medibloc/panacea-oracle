package key_test

import (
	"context"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/key"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/require"
)

// TestRetrieveAndStoreOraclePrivKey tests for a normal situation.
func TestRetrieveAndStoreOraclePrivKey(t *testing.T) {
	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	svc := &mocks.MockService{
		Config: config.DefaultConfig(),
		Sgx:    &mocks.MockSGX{},
		QueryClient: &mocks.MockQueryClient{
			OraclePubKey: oraclePrivKey.PubKey(),
		},
	}

	err = svc.Sgx.SealToFile(nodePrivKey.Serialize(), svc.Config.AbsNodePrivKeyPath())
	require.NoError(t, err)
	defer func() {
		os.Remove(svc.Config.AbsNodePrivKeyPath())
		os.Remove(svc.Config.AbsOraclePrivKeyPath())
	}()

	secretKey := crypto.DeriveSharedKey(oraclePrivKey, nodePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(secretKey, nil, oraclePrivKey.Serialize())
	require.NoError(t, err)

	err = key.RetrieveAndStoreOraclePrivKey(context.Background(), svc, encryptedOraclePrivKey)
	require.NoError(t, err)

	storedOraclePrivKeyBz, err := svc.Sgx.UnsealFromFile(svc.GetConfig().AbsOraclePrivKeyPath())
	require.NoError(t, err)

	require.Equal(t, oraclePrivKey.Serialize(), storedOraclePrivKeyBz)
}

// TestRetrieveAndStoreOraclePrivKeyExistOraclePrivKey tests that the OraclePrivKey exists and fails.
func TestRetrieveAndStoreOraclePrivKeyExistOraclePrivKey(t *testing.T) {
	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	svc := &mocks.MockService{
		Config: config.DefaultConfig(),
		Sgx:    &mocks.MockSGX{},
		QueryClient: &mocks.MockQueryClient{
			OraclePubKey: oraclePrivKey.PubKey(),
		},
	}

	err = svc.Sgx.SealToFile(nodePrivKey.Serialize(), svc.Config.AbsNodePrivKeyPath())
	require.NoError(t, err)
	err = svc.Sgx.SealToFile(oraclePrivKey.Serialize(), svc.Config.AbsOraclePrivKeyPath())
	require.NoError(t, err)
	defer func() {
		os.Remove(svc.Config.AbsNodePrivKeyPath())
		os.Remove(svc.Config.AbsOraclePrivKeyPath())
	}()

	secretKey := crypto.DeriveSharedKey(oraclePrivKey, nodePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(secretKey, nil, oraclePrivKey.Serialize())
	require.NoError(t, err)

	err = key.RetrieveAndStoreOraclePrivKey(context.Background(), svc, encryptedOraclePrivKey)
	require.ErrorContains(t, err, "the oracle private key already exists")
}

// AAA tests that the NodePrivKey fails because it doesn't exist.
func TestRetrieveAndStoreOraclePrivKeyNotExistNodePrivKey(t *testing.T) {
	nodePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	svc := &mocks.MockService{
		Config: config.DefaultConfig(),
		Sgx:    &mocks.MockSGX{},
		QueryClient: &mocks.MockQueryClient{
			OraclePubKey: oraclePrivKey.PubKey(),
		},
	}

	secretKey := crypto.DeriveSharedKey(oraclePrivKey, nodePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(secretKey, nil, oraclePrivKey.Serialize())
	require.NoError(t, err)

	err = key.RetrieveAndStoreOraclePrivKey(context.Background(), svc, encryptedOraclePrivKey)
	require.ErrorContains(t, err, "the node private key is not exists")
}

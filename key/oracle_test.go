package key_test

import (
	"context"
	"os"
	"testing"

	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/integration/suite"
	"github.com/medibloc/panacea-oracle/key"
	"github.com/medibloc/panacea-oracle/mocks"
)

type oracleTestSuite struct {
	mocks.MockTestSuite
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, &oracleTestSuite{})
}

func (suite *oracleTestSuite) BeforeTest(_, _ string) {
	suite.Initialize()
}

func (suite *oracleTestSuite) AfterTest(_, _ string) {
	os.Remove(suite.Config.AbsNodePrivKeyPath())
	os.Remove(suite.Config.AbsOraclePrivKeyPath())
}

// TestRetrieveAndStoreOraclePrivKey tests for a normal situation.
func (suite *oracleTestSuite) TestDecryptAndStoreOraclePrivKey() {
	suite.QueryClient.OraclePubKey = suite.OraclePubKey

	err := suite.SGX.SealToFile(suite.NodePrivKey.Serialize(), suite.Config.AbsNodePrivKeyPath())
	suite.Require().NoError(err)

	secretKey := crypto.DeriveSharedKey(suite.OraclePrivKey, suite.NodePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(secretKey, nil, suite.OraclePrivKey.Serialize())
	suite.Require().NoError(err)

	err = key.DecryptAndStoreOraclePrivKey(context.Background(), suite.Svc, encryptedOraclePrivKey)
	suite.Require().NoError(err)

	storedOraclePrivKeyBz, err := suite.SGX.UnsealFromFile(suite.Config.AbsOraclePrivKeyPath())
	suite.Require().NoError(err)

	suite.Require().Equal(suite.OraclePrivKey.Serialize(), storedOraclePrivKeyBz)
}

// TestRetrieveAndStoreOraclePrivKeyExistOraclePrivKey tests that the OraclePrivKey exists and fails.
func (suite *oracleTestSuite) TestRetrieveAndStoreOraclePrivKeyExistOraclePrivKey() {
	suite.OraclePubKey = suite.OraclePrivKey.PubKey()

	err := suite.SGX.SealToFile(suite.NodePrivKey.Serialize(), suite.Config.AbsNodePrivKeyPath())
	suite.Require().NoError(err)

	err = suite.SGX.SealToFile(suite.OraclePrivKey.Serialize(), suite.Config.AbsOraclePrivKeyPath())
	suite.Require().NoError(err)

	secretKey := crypto.DeriveSharedKey(suite.OraclePrivKey, suite.NodePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(secretKey, nil, suite.OraclePrivKey.Serialize())
	suite.Require().NoError(err)

	err = key.DecryptAndStoreOraclePrivKey(context.Background(), suite.Svc, encryptedOraclePrivKey)
	suite.Require().ErrorContains(err, "the oracle private key already exists")
}

// AAA tests that the NodePrivKey fails because it doesn't exist.
func (suite *oracleTestSuite) TestRetrieveAndStoreOraclePrivKeyNotExistNodePrivKey() {
	suite.QueryClient.OraclePubKey = suite.OraclePubKey

	secretKey := crypto.DeriveSharedKey(suite.OraclePrivKey, suite.NodePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(secretKey, nil, suite.OraclePrivKey.Serialize())
	suite.Require().NoError(err)

	err = key.DecryptAndStoreOraclePrivKey(context.Background(), suite.Svc, encryptedOraclePrivKey)
	suite.Require().ErrorContains(err, "the node private key is not exists")
}

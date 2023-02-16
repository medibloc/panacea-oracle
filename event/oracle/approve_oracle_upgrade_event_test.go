package oracle_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/config"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/suite"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type approveOracleUpgradeTestSuite struct {
	suite.Suite

	svc     *mocks.MockService
	errChan chan error

	productID []byte
	uniqueID  []byte

	oracleAcc     *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey
	nodePrivKey   *btcec.PrivateKey
}

func TestApproveOracleUpgradeTestSuite(t *testing.T) {
	suite.Run(t, &approveOracleUpgradeTestSuite{})
}

func (suite *approveOracleUpgradeTestSuite) BeforeTest(_, _ string) {
	mnemonic, _ := crypto.NewMnemonic()

	suite.errChan = make(chan error)
	suite.productID = []byte("productID")
	suite.uniqueID = []byte("uniqueID")
	suite.oracleAcc, _ = panacea.NewOracleAccount(mnemonic, 0, 0)
	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.svc = &mocks.MockService{
		QueryClient: &mocks.MockQueryClient{
			OraclePubKey: suite.oraclePrivKey.PubKey(),
			OracleUpgrade: &oracletypes.OracleUpgrade{
				UniqueId:      hex.EncodeToString(suite.uniqueID),
				OracleAddress: suite.oracleAcc.GetAddress(),
			},
		},
		OracleAccount: suite.oracleAcc,
		Sgx:           &mocks.MockSGX{},
		Config:        config.DefaultConfig(),
		EnclaveInfo:   sgx.NewEnclaveInfo(suite.productID, suite.uniqueID),
	}
}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *approveOracleUpgradeTestSuite) AfterTest(_, _ string) {
	os.Remove(suite.svc.GetConfig().AbsNodePrivKeyPath())
	os.Remove(suite.svc.GetConfig().AbsOraclePrivKeyPath())
}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *approveOracleUpgradeTestSuite) TestNameAndGetEventQuery() {
	e := oracle.NewApproveOracleUpgradeEvent(suite.svc, suite.errChan)

	suite.Require().Equal("ApproveOracleUpgradeEvent", e.Name())
	suite.Require().Contains(e.GetEventQuery(), "message.action = 'ApproveOracleUpgrade'")
	suite.Require().Contains(
		e.GetEventQuery(),
		fmt.Sprintf("%s.%s = '%s'",
			oracletypes.EventTypeApproveOracleUpgrade,
			oracletypes.AttributeKeyOracleAddress,
			suite.oracleAcc.GetAddress(),
		),
	)
	suite.Require().Contains(
		e.GetEventQuery(),
		fmt.Sprintf("%s.%s = '%s'",
			oracletypes.EventTypeApproveOracleUpgrade,
			oracletypes.AttributeKeyUniqueID,
			hex.EncodeToString(suite.uniqueID),
		),
	)
}

// TestEventHandler tests that the EventHandler function behavior succeeds.
func (suite *approveOracleUpgradeTestSuite) TestEventHandler() {
	svc := suite.svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleUpgradeEvent(svc, errChan)
	conf := svc.GetConfig()

	err := svc.GetSgx().SealToFile(suite.nodePrivKey.Serialize(), conf.NodePrivKeyFile)
	suite.Require().NoError(err)

	nodePubKey := suite.nodePrivKey.PubKey()
	sharedKey := crypto.DeriveSharedKey(suite.oraclePrivKey, nodePubKey, crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(sharedKey, nil, suite.oraclePrivKey.Serialize())
	suite.Require().NoError(err)

	suite.svc.QueryClient.OracleUpgrade.EncryptedOraclePrivKey = encryptedOraclePrivKey

	go func() {
		err := e.EventHandler(context.Background(), coretypes.ResultEvent{})
		suite.Require().NoError(err)
	}()

	err = <-errChan
	suite.Require().NoError(err)

	savedOraclePrivKey, err := os.ReadFile(conf.AbsOraclePrivKeyPath())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.oraclePrivKey.Serialize(), savedOraclePrivKey)
}

// TestEventHandlerExistOraclePrivKey tests that the OraclePrivKey exists and fails
func (suite *approveOracleUpgradeTestSuite) TestEventHandlerExistOraclePrivKey() {
	svc := suite.svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleUpgradeEvent(svc, errChan)
	conf := svc.GetConfig()
	err := os.WriteFile(conf.AbsOraclePrivKeyPath(), suite.oraclePrivKey.Serialize(), fs.ModePerm)
	suite.Require().NoError(err)

	go func() {
		err := e.EventHandler(context.Background(), coretypes.ResultEvent{})
		suite.Require().NoError(err)
	}()

	err = <-errChan
	suite.Require().ErrorContains(err, "the oracle private key already exists")
}

// TestEventHandlerNotExistNodePrivKey tests for a NodePrivKey that fails because it doesn't exist.
func (suite *approveOracleUpgradeTestSuite) TestEventHandlerNotExistNodePrivKey() {
	svc := suite.svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleUpgradeEvent(svc, errChan)

	go func() {
		err := e.EventHandler(context.Background(), coretypes.ResultEvent{})
		suite.Require().NoError(err)
	}()

	err := <-errChan
	suite.Require().ErrorContains(err, "the node private key is not exists")
}

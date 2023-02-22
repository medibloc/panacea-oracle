package oracle_test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"testing"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/mocks"
	"github.com/stretchr/testify/suite"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type approveOracleUpgradeTestSuite struct {
	mocks.MockTestSuite

	errChan chan error
}

func TestApproveOracleUpgradeTestSuite(t *testing.T) {
	suite.Run(t, &approveOracleUpgradeTestSuite{})
}

func (suite *approveOracleUpgradeTestSuite) BeforeTest(_, _ string) {
	suite.Initialize()

	suite.errChan = make(chan error)
	suite.QueryClient.OraclePubKey = suite.OraclePubKey
	suite.QueryClient.OracleUpgrade = &oracletypes.OracleUpgrade{
		UniqueId:      suite.UniqueID,
		OracleAddress: suite.OracleAcc.GetAddress(),
	}
}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *approveOracleUpgradeTestSuite) AfterTest(_, _ string) {
	os.Remove(suite.Config.AbsNodePrivKeyPath())
	os.Remove(suite.Config.AbsOraclePrivKeyPath())
}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *approveOracleUpgradeTestSuite) TestNameAndGetEventQuery() {
	e := oracle.NewApproveOracleUpgradeEvent(suite.Svc, suite.errChan)

	suite.Require().Equal("ApproveOracleUpgradeEvent", e.Name())
	suite.Require().Contains(e.GetEventQuery(), "message.action = 'ApproveOracleUpgrade'")
	suite.Require().Contains(
		e.GetEventQuery(),
		fmt.Sprintf("%s.%s = '%s'",
			oracletypes.EventTypeApproveOracleUpgrade,
			oracletypes.AttributeKeyOracleAddress,
			suite.OracleAcc.GetAddress(),
		),
	)
	suite.Require().Contains(
		e.GetEventQuery(),
		fmt.Sprintf("%s.%s = '%s'",
			oracletypes.EventTypeApproveOracleUpgrade,
			oracletypes.AttributeKeyUniqueID,
			suite.UniqueID,
		),
	)
}

// TestEventHandler tests that the EventHandler function behavior succeeds.
func (suite *approveOracleUpgradeTestSuite) TestEventHandler() {
	svc := suite.Svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleUpgradeEvent(svc, errChan)
	conf := suite.Config

	err := suite.SGX.SealToFile(suite.NodePrivKey.Serialize(), conf.NodePrivKeyFile)
	suite.Require().NoError(err)

	nodePubKey := suite.NodePrivKey.PubKey()
	sharedKey := crypto.DeriveSharedKey(suite.OraclePrivKey, nodePubKey, crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(sharedKey, nil, suite.OraclePrivKey.Serialize())
	suite.Require().NoError(err)

	suite.QueryClient.OracleUpgrade.EncryptedOraclePrivKey = encryptedOraclePrivKey

	go func() {
		err := e.EventHandler(context.Background(), coretypes.ResultEvent{})
		suite.Require().NoError(err)
	}()

	err = <-errChan
	suite.Require().NoError(err)

	savedOraclePrivKey, err := os.ReadFile(conf.AbsOraclePrivKeyPath())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.OraclePrivKey.Serialize(), savedOraclePrivKey)
}

// TestEventHandlerExistOraclePrivKey tests that the OraclePrivKey exists and fails
func (suite *approveOracleUpgradeTestSuite) TestEventHandlerExistOraclePrivKey() {
	svc := suite.Svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleUpgradeEvent(svc, errChan)
	conf := suite.Config
	err := os.WriteFile(conf.AbsOraclePrivKeyPath(), suite.OraclePrivKey.Serialize(), fs.ModePerm)
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
	svc := suite.Svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleUpgradeEvent(svc, errChan)

	go func() {
		err := e.EventHandler(context.Background(), coretypes.ResultEvent{})
		suite.Require().NoError(err)
	}()

	err := <-errChan
	suite.Require().ErrorContains(err, "the node private key is not exists")
}

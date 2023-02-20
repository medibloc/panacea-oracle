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

type approveOracleRegistrationTestSuite struct {
	mocks.MockTestSuite

	errChan chan error
}

func TestApproveOracleRegistrationTestSuite(t *testing.T) {
	suite.Run(t, &approveOracleRegistrationTestSuite{})
}

func (suite *approveOracleRegistrationTestSuite) BeforeTest(_, _ string) {
	suite.Initialize()

	suite.errChan = make(chan error)
	suite.QueryClient.OracleRegistration = &oracletypes.OracleRegistration{}
	suite.QueryClient.OraclePubKey = suite.OraclePubKey

}

func (suite *approveOracleRegistrationTestSuite) AfterTest(_, _ string) {
	os.Remove(suite.Svc.Config().AbsNodePrivKeyPath())
	os.Remove(suite.Svc.Config().AbsOraclePrivKeyPath())
}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *approveOracleRegistrationTestSuite) TestNameAndGetEventQuery() {
	e := oracle.NewApproveOracleRegistrationEvent(suite.Svc, suite.errChan)

	suite.Require().Equal("ApproveOracleRegistrationEvent", e.Name())
	suite.Require().Contains(e.GetEventQuery(), "message.action = 'ApproveOracleRegistration'")
	suite.Require().Contains(
		e.GetEventQuery(),
		fmt.Sprintf("%s.%s = '%s'",
			oracletypes.EventTypeApproveOracleRegistration,
			oracletypes.AttributeKeyOracleAddress,
			suite.OracleAcc.GetAddress(),
		),
	)
	suite.Require().Contains(
		e.GetEventQuery(),
		fmt.Sprintf("%s.%s = '%s'",
			oracletypes.EventTypeApproveOracleRegistration,
			oracletypes.AttributeKeyUniqueID,
			suite.UniqueID,
		),
	)
}

// TestEventHandler tests that the EventHandler function behavior succeeds.
func (suite *approveOracleRegistrationTestSuite) TestEventHandler() {
	svc := suite.Svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleRegistrationEvent(svc, errChan)
	conf := suite.Config

	err := suite.SGX.SealToFile(suite.NodePrivKey.Serialize(), conf.NodePrivKeyFile)
	suite.Require().NoError(err)

	sharedKey := crypto.DeriveSharedKey(suite.OraclePrivKey, suite.NodePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.Encrypt(sharedKey, nil, suite.OraclePrivKey.Serialize())
	suite.Require().NoError(err)

	suite.QueryClient.OracleRegistration = &oracletypes.OracleRegistration{
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

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
func (suite *approveOracleRegistrationTestSuite) TestEventHandlerExistOraclePrivKey() {
	svc := suite.Svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleRegistrationEvent(svc, errChan)
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
func (suite *approveOracleRegistrationTestSuite) TestEventHandlerNotExistNodePrivKey() {
	svc := suite.Svc
	errChan := suite.errChan

	e := oracle.NewApproveOracleRegistrationEvent(svc, errChan)

	go func() {
		err := e.EventHandler(context.Background(), coretypes.ResultEvent{})
		suite.Require().NoError(err)
	}()

	err := <-errChan
	suite.Require().ErrorContains(err, "the node private key is not exists")
}

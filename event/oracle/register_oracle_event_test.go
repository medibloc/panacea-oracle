package oracle_test

import (
	"context"
	"fmt"
	"testing"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/mocks"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/stretchr/testify/suite"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type registerOracleEventTestSuite struct {
	mocks.MockTestSuite

	targetOracleAcc *panacea.OracleAccount
}

func TestRegisterOracleEventTestSuite(t *testing.T) {
	suite.Run(t, &registerOracleEventTestSuite{})
}

func (suite *registerOracleEventTestSuite) BeforeTest(_, _ string) {
	suite.Initialize()

	targetMnemonic, _ := crypto.NewMnemonic()

	suite.targetOracleAcc, _ = panacea.NewOracleAccount(targetMnemonic, 0, 0)
	suite.QueryClient.OracleRegistration = &oracletypes.OracleRegistration{}
}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *registerOracleEventTestSuite) TestNameAndGetEventQuery() {
	e := oracle.NewRegisterOracleEvent(suite.Svc)

	suite.Require().Equal("RegisterOracleEvent", e.Name())
	suite.Require().Contains(e.GetEventQuery(), "message.action = 'RegisterOracle'")
}

// TestEventHandler tests that the EventHandler function behavior succeeds.
func (suite *registerOracleEventTestSuite) TestEventHandler() {
	e := oracle.NewRegisterOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.OracleRegistration.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().NoError(err)

	// Broadcast Msg Validation
	txMsgs := suite.Svc.BroadCastTxMsgs()
	suite.Require().Len(txMsgs, 1)
	txMsg := txMsgs[0].(*oracletypes.MsgApproveOracleRegistration)
	approvalMsg := txMsg.ApprovalSharingOracleKey
	suite.Require().Equal(suite.UniqueID, approvalMsg.ApproverUniqueId)
	suite.Require().Equal(suite.UniqueID, approvalMsg.TargetUniqueId)
	suite.Require().Equal(suite.OracleAcc.GetAddress(), approvalMsg.ApproverOracleAddress)
	suite.Require().Equal(suite.targetOracleAcc.GetAddress(), approvalMsg.TargetOracleAddress)

	// Decrypt OraclePrivateKey encrypted with NodePrivateKey and OraclePublicKey
	sharedKey := crypto.DeriveSharedKey(suite.NodePrivKey, suite.OraclePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey := approvalMsg.EncryptedOraclePrivKey
	decryptedOraclePrivKey, err := crypto.Decrypt(sharedKey, nil, encryptedOraclePrivKey)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.OraclePrivKey.Serialize(), decryptedOraclePrivKey)
}

// TestEventHandlerNotSameUniqueID tests for situations where the UniqueID fetched from the event is different from the UniqueID fetched from Panacea.
func (suite *registerOracleEventTestSuite) TestEventHandlerNotSameUniqueID() {
	e := oracle.NewRegisterOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{"invalid_unique_id"}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.OracleRegistration.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "requester's unique ID is different from this binary's unique ID.")
}

// TestEventHandlerInvalidTrustedBlockInfo tests for invalid block info in the OracleRegistration registered with Panacea.
func (suite *registerOracleEventTestSuite) TestEventHandlerInvalidTrustedBlockInfo() {
	e := oracle.NewRegisterOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.VerifyTrustedBlockInfoError = fmt.Errorf("invalid block")
	suite.QueryClient.OracleRegistration.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify trusted block information.")
}

// TestEventHandlerInvalidVerifyRemoteReport tests for invalid remote reports.
func (suite *registerOracleEventTestSuite) TestEventHandlerInvalidVerifyRemoteReport() {
	e := oracle.NewRegisterOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.SGX.VerifyRemoteReportError = fmt.Errorf("invalid remote report")
	suite.QueryClient.OracleRegistration.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify remote report")
}

// TestEventHandlerFailedBroadcast tests Panacea with a transaction that fails.
func (suite *registerOracleEventTestSuite) TestEventHandlerFailedBroadcast() {
	e := oracle.NewRegisterOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.OracleRegistration.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.Svc.SetBroadcastTxResponse(-1, "", fmt.Errorf("failed to broadcast"))

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to ApproveOracleRegistration transaction for new oracle registration")
}

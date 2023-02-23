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

type upgradeOracleEventTestSuite struct {
	mocks.MockTestSuite

	targetOracleAcc *panacea.OracleAccount
}

func TestUpgradeOracleEventTestSuite(t *testing.T) {
	suite.Run(t, &upgradeOracleEventTestSuite{})
}

func (suite *upgradeOracleEventTestSuite) BeforeTest(_, _ string) {
	suite.Initialize()

	targetMnemonic, _ := crypto.NewMnemonic()

	suite.targetOracleAcc, _ = panacea.NewOracleAccount(targetMnemonic, 0, 0)
	suite.QueryClient.Oracle = &oracletypes.Oracle{}
	suite.QueryClient.OracleUpgrade = &oracletypes.OracleUpgrade{}
	suite.QueryClient.OracleUpgradeInfo = &oracletypes.OracleUpgradeInfo{}

}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *upgradeOracleEventTestSuite) TestNameAndGetEventQuery() {
	e := oracle.NewUpgradeOracleEvent(suite.Svc)

	suite.Require().Equal("UpgradeOracleEvent", e.Name())
	suite.Require().Contains(e.GetEventQuery(), "message.action = 'UpgradeOracle'")
}

// TestEventHandler tests that the EventHandler function behavior succeeds.
func (suite *upgradeOracleEventTestSuite) TestEventHandler() {
	e := oracle.NewUpgradeOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.OracleUpgrade.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.QueryClient.OracleUpgradeInfo.UniqueId = suite.UniqueID
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().NoError(err)

	// Broadcast Msg Validation
	txMsgs := suite.Svc.BroadCastTxMsgs()
	suite.Require().Len(txMsgs, 1)
	txMsg := txMsgs[0].(*oracletypes.MsgApproveOracleUpgrade)
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
func (suite *upgradeOracleEventTestSuite) TestEventHandlerNotSameUniqueID() {
	e := oracle.NewUpgradeOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{"invalid_unique_id"}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.OracleUpgrade.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.QueryClient.OracleUpgradeInfo.UniqueId = suite.UniqueID
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "the upgrade unique ID is different from the one stored in panacea")
}

// TestEventHandlerInvalidTrustedBlockInfo tests for invalid block info in the OracleRegistration registered with Panacea.
func (suite *upgradeOracleEventTestSuite) TestEventHandlerInvalidTrustedBlockInfo() {
	e := oracle.NewUpgradeOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.VerifyTrustedBlockInfoError = fmt.Errorf("invalid block")
	suite.QueryClient.OracleUpgrade.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.QueryClient.OracleUpgradeInfo.UniqueId = suite.UniqueID
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify trusted block information")
}

// TestEventHandlerInvalidVerifyRemoteReport tests for invalid remote reports.
func (suite *upgradeOracleEventTestSuite) TestEventHandlerInvalidVerifyRemoteReport() {
	e := oracle.NewUpgradeOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.SGX.VerifyRemoteReportError = fmt.Errorf("invalid trusted block info")
	suite.QueryClient.OracleUpgrade.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.QueryClient.OracleUpgradeInfo.UniqueId = suite.UniqueID
	suite.Svc.SetBroadcastTxResponse(0, "", nil)

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify remote report")
}

// TestEventHandlerFailedBroadcast tests Panacea with a transaction that fails.
func (suite *upgradeOracleEventTestSuite) TestEventHandlerFailedBroadcast() {
	e := oracle.NewUpgradeOracleEvent(suite.Svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{suite.UniqueID}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.QueryClient.OracleUpgrade.NodePubKey = suite.NodePrivKey.PubKey().SerializeCompressed()
	suite.QueryClient.OracleUpgradeInfo.UniqueId = suite.UniqueID
	suite.Svc.SetBroadcastTxResponse(-1, "", fmt.Errorf("failed to broadcast"))

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to ApproveOracleUpgrade transaction for oracle upgrade")
}

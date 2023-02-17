package oracle_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/crypto"
	"github.com/medibloc/panacea-oracle/event/oracle"
	"github.com/medibloc/panacea-oracle/mocks"
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/stretchr/testify/suite"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type upgradeOracleEventTestSuite struct {
	suite.Suite

	svc *mocks.MockService

	uniqueID []byte

	approvalOracleAcc *panacea.OracleAccount
	targetOracleAcc   *panacea.OracleAccount
	oraclePrivKey     *btcec.PrivateKey
	nodePrivKey       *btcec.PrivateKey
}

func TestUpgradeOracleEventTestSuite(t *testing.T) {
	suite.Run(t, &upgradeOracleEventTestSuite{})
}

func (suite *upgradeOracleEventTestSuite) BeforeTest(_, _ string) {
	approvalMnemonic, _ := crypto.NewMnemonic()
	targetMnemonic, _ := crypto.NewMnemonic()

	suite.uniqueID = []byte("uniqueID")
	suite.approvalOracleAcc, _ = panacea.NewOracleAccount(approvalMnemonic, 0, 0)
	suite.targetOracleAcc, _ = panacea.NewOracleAccount(targetMnemonic, 0, 0)
	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())

	suite.svc = &mocks.MockService{
		QueryClient: &mocks.MockQueryClient{
			Oracle:            &oracletypes.Oracle{},
			OracleUpgrade:     &oracletypes.OracleUpgrade{},
			OracleUpgradeInfo: &oracletypes.OracleUpgradeInfo{},
		},
		Sgx:           &mocks.MockSGX{},
		EnclaveInfo:   sgx.NewEnclaveInfo(nil, suite.uniqueID),
		OracleAccount: suite.approvalOracleAcc,
		OraclePrivKey: suite.oraclePrivKey,
	}
}

// TestNameAndGetEventQuery tests the name and eventQuery.
func (suite *upgradeOracleEventTestSuite) TestNameAndGetEventQuery() {
	e := oracle.NewUpgradeOracleEvent(suite.svc)

	suite.Require().Equal("UpgradeOracleEvent", e.Name())
	suite.Require().Contains(e.GetEventQuery(), "message.action = 'UpgradeOracle'")
}

// TestEventHandler tests that the EventHandler function behavior succeeds.
func (suite *upgradeOracleEventTestSuite) TestEventHandler() {
	e := oracle.NewUpgradeOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.OracleUpgrade.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.QueryClient.OracleUpgradeInfo.UniqueId = hex.EncodeToString(suite.uniqueID)
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().NoError(err)

	// Broadcast Msg Validation
	txMsgs := suite.svc.GetBroadCastTxMsgs()
	suite.Require().Len(txMsgs, 1)
	txMsg := txMsgs[0].(*oracletypes.MsgApproveOracleUpgrade)
	approvalMsg := txMsg.ApprovalSharingOracleKey
	suite.Require().Equal(hex.EncodeToString(suite.uniqueID), approvalMsg.ApproverUniqueId)
	suite.Require().Equal(hex.EncodeToString(suite.uniqueID), approvalMsg.TargetUniqueId)
	suite.Require().Equal(suite.approvalOracleAcc.GetAddress(), approvalMsg.ApproverOracleAddress)
	suite.Require().Equal(suite.targetOracleAcc.GetAddress(), approvalMsg.TargetOracleAddress)

	// Decrypt OraclePrivateKey encrypted with NodePrivateKey and OraclePublicKey
	sharedKey := crypto.DeriveSharedKey(suite.nodePrivKey, suite.oraclePrivKey.PubKey(), crypto.KDFSHA256)
	encryptedOraclePrivKey := approvalMsg.EncryptedOraclePrivKey
	decryptedOraclePrivKey, err := crypto.Decrypt(sharedKey, nil, encryptedOraclePrivKey)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.oraclePrivKey.Serialize(), decryptedOraclePrivKey)
}

// TestEventHandlerNotSameUniqueID tests for situations where the UniqueID fetched from the event is different from the UniqueID fetched from Panacea.
func (suite *upgradeOracleEventTestSuite) TestEventHandlerNotSameUniqueID() {
	e := oracle.NewUpgradeOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{"invalid_unique_id"}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.OracleUpgrade.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.QueryClient.OracleUpgradeInfo.UniqueId = hex.EncodeToString(suite.uniqueID)
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "the upgrade unique ID is different from the one stored in panacea")
}

// TestEventHandlerInvalidTrustedBlockInfo tests for invalid block info in the OracleRegistration registered with Panacea.
func (suite *upgradeOracleEventTestSuite) TestEventHandlerInvalidTrustedBlockInfo() {
	e := oracle.NewUpgradeOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.VerifyTrustedBlockInfoError = fmt.Errorf("invalid block")
	suite.svc.QueryClient.OracleUpgrade.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.QueryClient.OracleUpgradeInfo.UniqueId = hex.EncodeToString(suite.uniqueID)
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify trusted block information")
}

// TestEventHandlerInvalidVerifyRemoteReport tests for invalid remote reports.
func (suite *upgradeOracleEventTestSuite) TestEventHandlerInvalidVerifyRemoteReport() {
	e := oracle.NewUpgradeOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.Sgx.VerifyRemoteReportError = fmt.Errorf("invalid trusted block info")
	suite.svc.QueryClient.OracleUpgrade.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.QueryClient.OracleUpgradeInfo.UniqueId = hex.EncodeToString(suite.uniqueID)
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify remote report")
}

// TestEventHandlerFailedBroadcast tests Panacea with a transaction that fails.
func (suite *upgradeOracleEventTestSuite) TestEventHandlerFailedBroadcast() {
	e := oracle.NewUpgradeOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.OracleUpgrade.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.QueryClient.OracleUpgradeInfo.UniqueId = hex.EncodeToString(suite.uniqueID)
	suite.svc.BroadcastCode = -1
	suite.svc.BroadcastError = fmt.Errorf("failed to broadcast")

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to ApproveOracleUpgrade transaction for oracle upgrade")
}

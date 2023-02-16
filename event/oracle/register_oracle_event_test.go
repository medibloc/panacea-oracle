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
	"github.com/medibloc/panacea-oracle/panacea"
	"github.com/medibloc/panacea-oracle/sgx"
	"github.com/medibloc/panacea-oracle/types/test_utils/mocks"
	"github.com/stretchr/testify/suite"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type registerOracleEventTestSuite struct {
	suite.Suite

	svc *mocks.MockService

	uniqueID []byte

	approvalOracleAcc *panacea.OracleAccount
	targetOracleAcc   *panacea.OracleAccount
	oraclePrivKey     *btcec.PrivateKey
	nodePrivKey       *btcec.PrivateKey
}

func TestRegisterOracleEventTestSuite(t *testing.T) {
	suite.Run(t, &registerOracleEventTestSuite{})
}

func (suite *registerOracleEventTestSuite) BeforeTest(_, _ string) {
	approvalMnemonic, _ := crypto.NewMnemonic()
	targetMnemonic, _ := crypto.NewMnemonic()

	suite.uniqueID = []byte("uniqueID")
	suite.approvalOracleAcc, _ = panacea.NewOracleAccount(approvalMnemonic, 0, 0)
	suite.targetOracleAcc, _ = panacea.NewOracleAccount(targetMnemonic, 0, 0)
	suite.oraclePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	suite.nodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())

	suite.svc = &mocks.MockService{
		QueryClient: &mocks.MockQueryClient{
			OracleRegistration: &oracletypes.OracleRegistration{},
		},
		Sgx:           &mocks.MockSGX{},
		EnclaveInfo:   sgx.NewEnclaveInfo(nil, suite.uniqueID),
		OracleAccount: suite.approvalOracleAcc,
		OraclePrivKey: suite.oraclePrivKey,
	}
}

func (suite *registerOracleEventTestSuite) TestNameAndGetEventQuery() {
	e := oracle.NewRegisterOracleEvent(suite.svc)

	suite.Require().Equal("RegisterOracleEvent", e.Name())
	suite.Require().Contains(e.GetEventQuery(), "message.action = 'RegisterOracle'")
}

func (suite *registerOracleEventTestSuite) TestEventHandler() {
	e := oracle.NewRegisterOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.OracleRegistration.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().NoError(err)

	// Broadcast Msg Validation
	txMsgs := suite.svc.GetBroadCastTxMsgs()
	suite.Require().Len(txMsgs, 1)
	txMsg := txMsgs[0].(*oracletypes.MsgApproveOracleRegistration)
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

func (suite *registerOracleEventTestSuite) TestEventHandlerNotSameUniqueID() {
	e := oracle.NewRegisterOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{"invalid_unique_id"}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.OracleRegistration.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "requester's unique ID is different from this binary's unique ID.")
}

func (suite *registerOracleEventTestSuite) TestEventHandlerInvalidTrustedBlockInfo() {
	e := oracle.NewRegisterOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.VerifyTrustedBlockInfoError = fmt.Errorf("invalid block")
	suite.svc.QueryClient.OracleRegistration.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify trusted block information.")
}

func (suite *registerOracleEventTestSuite) TestEventHandlerInvalidVerifyRemoteReport() {
	e := oracle.NewRegisterOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.Sgx.VerifyRemoteReportError = fmt.Errorf("invalid remote report")
	suite.svc.QueryClient.OracleRegistration.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.BroadcastCode = 0

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to verify remote report")
}

func (suite *registerOracleEventTestSuite) TestEventHandlerFailedBroadcast() {
	e := oracle.NewRegisterOracleEvent(suite.svc)

	events := make(map[string][]string)
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID] = []string{hex.EncodeToString(suite.uniqueID)}
	events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress] = []string{suite.targetOracleAcc.GetAddress()}

	resultEvent := coretypes.ResultEvent{
		Events: events,
	}

	suite.svc.QueryClient.OracleRegistration.NodePubKey = suite.nodePrivKey.PubKey().SerializeCompressed()
	suite.svc.BroadcastCode = -1
	suite.svc.BroadcastError = fmt.Errorf("failed to broadcast")

	err := e.EventHandler(context.Background(), resultEvent)
	suite.Require().ErrorContains(err, "failed to ApproveOracleRegistration transaction for new oracle registration")
}

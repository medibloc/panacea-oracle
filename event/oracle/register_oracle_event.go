package oracle

import (
	"context"
	"crypto/sha256"
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/event"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	reactor event.Service
}

func NewRegisterOracleEvent(s event.Service) RegisterOracleEvent {
	return RegisterOracleEvent{s}
}

func (e RegisterOracleEvent) Name() string {
	return "RegisterOracleEvent"
}

func (e RegisterOracleEvent) GetEventQuery() string {
	return "message.action = 'RegisterOracle'"
}

func (e RegisterOracleEvent) EventHandler(ctx context.Context, event ctypes.ResultEvent) error {
	uniqueID := event.Events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID][0]
	targetAddress := event.Events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress][0]

	// get oracle registration
	oracleRegistration, err := e.reactor.GetQueryClient().GetOracleRegistration(ctx, uniqueID, targetAddress)
	if err != nil {
		return fmt.Errorf("failed to get oracle registration. unique ID(%s), target address(%s): %w", uniqueID, targetAddress, err)
	}

	// verify oracle registration
	if err := e.verifyOracleRegistration(oracleRegistration, uniqueID); err != nil {
		return fmt.Errorf("failed to verify oracle registration. unique ID(%s), target address(%s): %w", uniqueID, targetAddress, err)
	}

	// generate Msg/ApproveOracleRegistration
	msgApproveOracleRegistration, err := e.generateApproveOracleRegistrationMsg(oracleRegistration, uniqueID, targetAddress)
	if err != nil {
		return fmt.Errorf("failed to generate MsgApproveOracleRegistration: %w", err)
	}

	log.Infof("new oracle registration approval info. unique ID(%s), approver address(%s), target address(%s)",
		msgApproveOracleRegistration.ApprovalSharingOracleKey.ApproverUniqueId,
		msgApproveOracleRegistration.ApprovalSharingOracleKey.ApproverOracleAddress,
		msgApproveOracleRegistration.ApprovalSharingOracleKey.TargetOracleAddress,
	)

	txHeight, txHash, err := e.reactor.BroadcastTx(msgApproveOracleRegistration)
	if err != nil {
		return fmt.Errorf("failed to ApproveOracleRegistration transaction for new oracle registration: %w", err)
	}

	log.Infof("succeeded to ApproveOracleRegistration transaction for new oracle registration. height(%d), hash(%s)", txHeight, txHash)

	return nil
}

func (e RegisterOracleEvent) verifyOracleRegistration(oracleRegistration *oracletypes.OracleRegistration, uniqueID string) error {
	queryClient := e.reactor.GetQueryClient()
	approverUniqueID := e.reactor.GetEnclaveInfo().UniqueIDHex()

	if uniqueID != approverUniqueID {
		return fmt.Errorf("requester's unique ID is different from this binary's unique ID. expected(%s) got(%s)", approverUniqueID, uniqueID)
	}

	if err := queryClient.VerifyTrustedBlockInfo(oracleRegistration.TrustedBlockHeight, oracleRegistration.TrustedBlockHash); err != nil {
		return fmt.Errorf("failed to verify trusted block information. height(%d), hash(%s): %w", oracleRegistration.TrustedBlockHeight, oracleRegistration.TrustedBlockHash, err)
	}

	nodePubKeyHash := sha256.Sum256(oracleRegistration.NodePubKey)

	if err := e.reactor.GetSgx().VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, nodePubKeyHash[:], e.reactor.GetEnclaveInfo().UniqueID); err != nil {
		return fmt.Errorf("failed to verify remote report: %w", err)
	}

	return nil
}

func (e RegisterOracleEvent) generateApproveOracleRegistrationMsg(oracleRegistration *oracletypes.OracleRegistration, targetUniqueID, targetAddress string) (*oracletypes.MsgApproveOracleRegistration, error) {
	approverAddress := e.reactor.GetOracleAcc().GetAddress()
	oraclePrivKeyBz := e.reactor.GetOraclePrivKey().Serialize()
	approverUniqueID := e.reactor.GetEnclaveInfo().UniqueIDHex()

	encryptedOraclePrivKey, err := encryptOraclePrivKey(oraclePrivKeyBz, oracleRegistration.NodePubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt oracle private key: %w", err)
	}

	approvalMsg := &oracletypes.ApprovalSharingOracleKey{
		ApproverUniqueId:       approverUniqueID,
		ApproverOracleAddress:  approverAddress,
		TargetUniqueId:         targetUniqueID,
		TargetOracleAddress:    targetAddress,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	sig, err := signApprovalMsg(approvalMsg, oraclePrivKeyBz)
	if err != nil {
		return nil, fmt.Errorf("failed to sign for approval of oracle registration: %w", err)
	}

	return &oracletypes.MsgApproveOracleRegistration{
		ApprovalSharingOracleKey: approvalMsg,
		Signature:                sig,
	}, nil
}

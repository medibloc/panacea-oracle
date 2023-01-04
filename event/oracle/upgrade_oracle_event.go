package oracle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*UpgradeOracleEvent)(nil)

type UpgradeOracleEvent struct {
	reactor event.Service
}

func NewUpgradeOracleEvent(s event.Service) UpgradeOracleEvent {
	return UpgradeOracleEvent{s}
}

func (e UpgradeOracleEvent) Name() string {
	return "UpgradeOracleEvent"
}

func (e UpgradeOracleEvent) GetEventQuery() string {
	return "message.action = 'UpgradeOracle'"
}

func (e UpgradeOracleEvent) EventHandler(ctx context.Context, event ctypes.ResultEvent) error {
	uniqueID := event.Events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyUniqueID][0]
	targetAddress := event.Events[oracletypes.EventTypeUpgrade+"."+oracletypes.AttributeKeyOracleAddress][0]

	// get oracle upgrade
	oracleUpgrade, err := e.reactor.QueryClient().GetOracleUpgrade(ctx, uniqueID, targetAddress)
	if err != nil {
		return fmt.Errorf("failed to get oracle upgrade. unique ID(%s), target address(%s): %w", uniqueID, targetAddress, err)
	}

	// verify oracle upgrade
	if err := e.verifyOracleUpgrade(ctx, oracleUpgrade, uniqueID, targetAddress); err != nil {
		return fmt.Errorf("failed to verify oracle upgrade. unique ID(%s), target address(%s): %w", uniqueID, targetAddress, err)
	}

	// generate Msg/ApproveOracleUpgrade
	msgApproveOracleUpgrade, err := e.generateApproveOracleUpgradeMsg(oracleUpgrade, uniqueID, targetAddress)
	if err != nil {
		return fmt.Errorf("failed to generate MsgApproveOracleUpgrade: %w", err)
	}

	log.Infof("oracle upgrade approval info. unique ID(%s), approver address(%s), target address(%s)",
		msgApproveOracleUpgrade.ApprovalSharingOracleKey.ApproverUniqueId,
		msgApproveOracleUpgrade.ApprovalSharingOracleKey.ApproverOracleAddress,
		msgApproveOracleUpgrade.ApprovalSharingOracleKey.TargetOracleAddress,
	)

	txHeight, txHash, err := e.reactor.BroadcastTx(msgApproveOracleUpgrade)
	if err != nil {
		return fmt.Errorf("failed to ApproveOracleUpgrade transaction for oracle upgrade: %w", err)
	}

	log.Infof("succeeded to ApproveOracleUpgrae transaction for oracle upgrade. height(%d), hash(%s)", txHeight, txHash)

	return nil
}

func (e UpgradeOracleEvent) verifyOracleUpgrade(ctx context.Context, oracleUpgrade *oracletypes.OracleUpgrade, uniqueID, targetAddress string) error {
	queryClient := e.reactor.QueryClient()

	oracleUpgradeInfo, err := queryClient.GetOracleUpgradeInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get oracle upgrade info: %w", err)
	}

	// check if the unique ID is same with the one stored in panacea
	if uniqueID != oracleUpgradeInfo.GetUniqueId() {
		return fmt.Errorf("the upgrade unique ID is different from the one stored in panacea")
	}

	// check if the oracle is registered
	if _, err := queryClient.GetOracle(ctx, targetAddress); err != nil {
		return fmt.Errorf("failed to get oracle: %w", err)
	}

	// verify trusted block info
	if err := queryClient.VerifyTrustedBlockInfo(oracleUpgrade.TrustedBlockHeight, oracleUpgrade.TrustedBlockHash); err != nil {
		return fmt.Errorf("failed to verify trusted block information. height(%d), hash(%s): %w", oracleUpgrade.TrustedBlockHeight, oracleUpgrade.TrustedBlockHash, err)
	}

	// verify remote report
	nodePubKeyHash := sha256.Sum256(oracleUpgrade.NodePubKey)

	uniqueIDBz, err := hex.DecodeString(oracleUpgradeInfo.GetUniqueId())
	if err != nil {
		return fmt.Errorf("failed to decode unique ID: %w", err)
	}

	if err := sgx.VerifyRemoteReport(oracleUpgrade.GetNodePubKeyRemoteReport(), nodePubKeyHash[:], uniqueIDBz); err != nil {
		return fmt.Errorf("failed to verify remote report: %w", err)
	}

	return nil
}

func (e UpgradeOracleEvent) generateApproveOracleUpgradeMsg(oracleUpgrade *oracletypes.OracleUpgrade, targetUniqueID, targetAddress string) (*oracletypes.MsgApproveOracleUpgrade, error) {
	approverAddress := e.reactor.OracleAcc().GetAddress()
	oraclePrivKeyBz := e.reactor.OraclePrivKey().Serialize()
	approverUniqueID := e.reactor.EnclaveInfo().UniqueIDHex()

	// generate transaction message for approval of oracle upgrade
	encryptedOraclePrivKey, err := encryptOraclePrivKey(oraclePrivKeyBz, oracleUpgrade.NodePubKey)
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
		return nil, fmt.Errorf("failed to sign for approval of oracle upgrade: %w", err)
	}

	return &oracletypes.MsgApproveOracleUpgrade{
		ApprovalSharingOracleKey: approvalMsg,
		Signature:                sig,
	}, nil
}

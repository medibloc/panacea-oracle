package oracle

import (
	"crypto/sha256"
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/sgx"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	service event.Service
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

func (e RegisterOracleEvent) EventHandler(resultEvent ctypes.ResultEvent) error {
	height, err := event.GetQueryHeight(e.service.QueryClient(), resultEvent)
	if err != nil {
		return err
	}

	uniqueID := resultEvent.Events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyUniqueID][0]
	targetAddress := resultEvent.Events[oracletypes.EventTypeRegistration+"."+oracletypes.AttributeKeyOracleAddress][0]

	msgApproveOracleRegistration, err := e.verifyAndGetMsgApproveOracleRegistration(height, uniqueID, targetAddress)
	if err != nil {
		return err
	}

	log.Infof("new oracle registration approval info. uniqueID(%s), approverAddress(%s), targetAddress(%s)",
		msgApproveOracleRegistration.ApproveOracleRegistration.UniqueId,
		msgApproveOracleRegistration.ApproveOracleRegistration.ApproverOracleAddress,
		msgApproveOracleRegistration.ApproveOracleRegistration.TargetOracleAddress,
	)

	txHeight, txHash, err := e.service.BroadcastTx(msgApproveOracleRegistration)
	if err != nil {
		return fmt.Errorf("failed to ApproveOracleRegistration transaction for new oracle registration: %v", err)
	}

	log.Infof("succeeded to ApproveOracleRegistration transaction for new oracle registration. height(%v), hash(%s)", txHeight, txHash)

	return nil
}

func (e RegisterOracleEvent) verifyAndGetMsgApproveOracleRegistration(height int64, uniqueID, targetAddress string) (*oracletypes.MsgApproveOracleRegistration, error) {
	queryClient := e.service.QueryClient()
	approverAddress := e.service.OracleAcc().GetAddress()
	oraclePrivKeyBz := e.service.OraclePrivKey().Serialize()
	approverUniqueID := e.service.EnclaveInfo().UniqueIDHex()

	if uniqueID != approverUniqueID {
		return nil, fmt.Errorf("oracle's uniqueID does not match the requested uniqueID. expected(%s) got(%s)", approverUniqueID, uniqueID)
	} else {
		oracleRegistration, err := queryClient.GetOracleRegistration(height, uniqueID, targetAddress)
		if err != nil {
			log.Errorf("err while get oracleRegistration: %v", err)
			return nil, err
		}

		if err := verifyTrustedBlockInfo(e.service.QueryClient(), oracleRegistration.TrustedBlockHeight, oracleRegistration.TrustedBlockHash); err != nil {
			log.Errorf("failed to verify trusted block. height(%d), hash(%s), err(%v)", oracleRegistration.TrustedBlockHeight, oracleRegistration.TrustedBlockHash, err)
			return nil, err
		}

		nodePubKeyHash := sha256.Sum256(oracleRegistration.NodePubKey)

		if err := sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, nodePubKeyHash[:], *e.service.EnclaveInfo()); err != nil {
			log.Errorf("failed to verification report. uniqueID(%s), address(%s), err(%v)", oracleRegistration.UniqueId, oracleRegistration.OracleAddress, err)
			return nil, err
		}

		return makeMsgApproveOracleRegistration(approverUniqueID, approverAddress, targetAddress, oraclePrivKeyBz, oracleRegistration.NodePubKey)
	}
}

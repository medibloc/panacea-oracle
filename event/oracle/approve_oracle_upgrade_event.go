package oracle

import (
	"context"
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/event"
	"github.com/medibloc/panacea-oracle/key"
	"github.com/medibloc/panacea-oracle/service"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*ApproveOracleUpgradeEvent)(nil)

type ApproveOracleUpgradeEvent struct {
	service  service.Service
	doneChan chan error
}

func NewApproveOracleUpgradeEvent(s service.Service, donChan chan error) ApproveOracleUpgradeEvent {
	return ApproveOracleUpgradeEvent{s, donChan}
}

func (e ApproveOracleUpgradeEvent) Name() string {
	return "ApproveOracleUpgradeEvent"
}

func (e ApproveOracleUpgradeEvent) GetEventQuery() string {
	return fmt.Sprintf("message.action = 'ApproveOracleUpgrade' and %s.%s = '%s' and %s.%s = '%s'",
		oracletypes.EventTypeApproveOracleUpgrade,
		oracletypes.AttributeKeyOracleAddress,
		e.service.GetOracleAcc().GetAddress(),
		oracletypes.EventTypeApproveOracleUpgrade,
		oracletypes.AttributeKeyUniqueID,
		e.service.GetEnclaveInfo().UniqueIDHex(),
	)
}

func (e ApproveOracleUpgradeEvent) EventHandler(ctx context.Context, _ ctypes.ResultEvent) error {
	uniqueID := e.service.GetEnclaveInfo().UniqueIDHex()
	oracleAddress := e.service.GetOracleAcc().GetAddress()
	oracleUpgrade, err := e.service.GetQueryClient().GetOracleUpgrade(ctx, uniqueID, oracleAddress)
	if err != nil {
		e.doneChan <- fmt.Errorf("failed to get oracle upgrade: %w", err)
	}

	e.doneChan <- key.RetrieveAndStoreOraclePrivKey(ctx, e.service, oracleUpgrade.EncryptedOraclePrivKey)

	return nil
}

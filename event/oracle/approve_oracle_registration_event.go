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

var _ event.Event = (*ApproveOracleRegistrationEvent)(nil)

type ApproveOracleRegistrationEvent struct {
	service  service.Service
	doneChan chan error
}

func NewApproveOracleRegistrationEvent(s service.Service, doneChan chan error) ApproveOracleRegistrationEvent {
	return ApproveOracleRegistrationEvent{s, doneChan}
}

func (e ApproveOracleRegistrationEvent) Name() string {
	return "ApproveOracleRegistrationEvent"
}

func (e ApproveOracleRegistrationEvent) GetEventQuery() string {
	return fmt.Sprintf("message.action = 'ApproveOracleRegistration' and %s.%s = '%s' and %s.%s = '%s'",
		oracletypes.EventTypeApproveOracleRegistration,
		oracletypes.AttributeKeyOracleAddress,
		e.service.GetOracleAcc().GetAddress(),
		oracletypes.EventTypeApproveOracleRegistration,
		oracletypes.AttributeKeyUniqueID,
		e.service.GetEnclaveInfo().UniqueIDHex(),
	)
}

func (e ApproveOracleRegistrationEvent) EventHandler(ctx context.Context, _ ctypes.ResultEvent) error {
	uniqueID := e.service.GetEnclaveInfo().UniqueIDHex()
	oracleAddress := e.service.GetOracleAcc().GetAddress()
	oracleRegistration, err := e.service.GetQueryClient().GetOracleRegistration(ctx, uniqueID, oracleAddress)
	if err != nil {
		e.doneChan <- fmt.Errorf("failed to get oracle registration: %w", err)
	}

	e.doneChan <- key.RetrieveAndStoreOraclePrivKey(ctx, e.service, oracleRegistration.EncryptedOraclePrivKey)

	return nil
}

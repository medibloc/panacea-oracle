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
		e.service.OracleAcc().GetAddress(),
		oracletypes.EventTypeApproveOracleRegistration,
		oracletypes.AttributeKeyUniqueID,
		e.service.EnclaveInfo().UniqueIDHex(),
	)
}

func (e ApproveOracleRegistrationEvent) EventHandler(ctx context.Context, _ ctypes.ResultEvent) error {
	uniqueID := e.service.EnclaveInfo().UniqueIDHex()
	oracleAddress := e.service.OracleAcc().GetAddress()
	oracleRegistration, err := e.service.QueryClient().GetOracleRegistration(ctx, uniqueID, oracleAddress)
	if err != nil {
		e.doneChan <- fmt.Errorf("failed to get oracle registration: %w", err)
	}

	e.doneChan <- key.DecryptAndStoreOraclePrivKey(ctx, e.service, oracleRegistration.EncryptedOraclePrivKey)

	return nil
}

package oracle

import (
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-oracle/event"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*ApproveOracleRegistrationEvent)(nil)

type ApproveOracleRegistrationEvent struct {
	reactor  event.OracleService
	doneChan chan error
}

func NewApproveOracleRegistrationEvent(s event.OracleService, doneChan chan error) ApproveOracleRegistrationEvent {
	return ApproveOracleRegistrationEvent{s, doneChan}
}

func (e ApproveOracleRegistrationEvent) Name() string {
	return "ApproveOracleRegistrationEvent"
}

func (e ApproveOracleRegistrationEvent) GetEventQuery() string {
	return fmt.Sprintf("message.action = 'ApproveOracleRegistration' and %s.%s = '%s' and %s.%s = '%s'",
		oracletypes.EventTypeApproveOracleRegistration,
		oracletypes.AttributeKeyOracleAddress,
		e.reactor.OracleAcc().GetAddress(),
		oracletypes.EventTypeApproveOracleRegistration,
		oracletypes.AttributeKeyUniqueID,
		e.reactor.EnclaveInfo().UniqueIDHex(),
	)
}

func (e ApproveOracleRegistrationEvent) EventHandler(event ctypes.ResultEvent) error {
	e.doneChan <- e.reactor.GetAndStoreOraclePrivKey()
	return nil
}

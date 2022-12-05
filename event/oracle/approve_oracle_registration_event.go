package oracle

import (
	"github.com/medibloc/panacea-oracle/event"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*ApproveOracleRegistrationEvent)(nil)

type ApproveOracleRegistrationEvent struct {
	reactor event.Reactor
}

func NewApproveOracleRegistrationEvent(s event.Reactor) ApproveOracleRegistrationEvent {
	return ApproveOracleRegistrationEvent{s}
}

func (e ApproveOracleRegistrationEvent) GetEventQuery() string {
	return "message.action = 'ApproveOracleRegistration'"
}

// TODO: EventHandler for ApproveOracleRegistration will be implemented when ApproveOracleRegistration Tx implemented in panacea-core.
func (e ApproveOracleRegistrationEvent) EventHandler(event ctypes.ResultEvent) error {
	panic("implement me")
}

package oracle

import (
	"github.com/medibloc/panacea-oracle/event"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*ApproveOracleRegistrationEvent)(nil)

type ApproveOracleRegistrationEvent struct {
	reactor event.Reactor
}

func NewApproveOracleRegistrationEvent(s event.Reactor) RegisterOracleEvent {
	return RegisterOracleEvent{s}
}

func (e ApproveOracleRegistrationEvent) GetEventQuery() string {
	return "message.action = 'ApproveOracleRegistration'"
}

func (e ApproveOracleRegistrationEvent) EventHandler(event ctypes.ResultEvent) error {
	panic("implement me")
}

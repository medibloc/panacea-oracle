package oracle

import (
	"github.com/medibloc/panacea-oracle/event"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*ApproveOracleRegistrationEvent)(nil)

type ApproveOracleRegistrationEvent struct {
}

func NewApproveOracleRegistrationEvent() ApproveOracleRegistrationEvent {
	return ApproveOracleRegistrationEvent{}
}

func (e ApproveOracleRegistrationEvent) GetEventQuery() string {
	return "message.action = 'ApproveOracleRegistration'"
}

func (e ApproveOracleRegistrationEvent) EventHandler(event ctypes.ResultEvent) error {
	// TODO: implement to retrieve and store oracle_priv_key
	panic("implement me")
}

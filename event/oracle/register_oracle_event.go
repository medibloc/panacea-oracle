package oracle

import (
	"github.com/medibloc/panacea-oracle/event"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	reactor event.Reactor
}

func NewRegisterOracleEvent(s event.Reactor) RegisterOracleEvent {
	return RegisterOracleEvent{s}
}

func (e RegisterOracleEvent) GetEventQuery() string {
	return "message.action = 'RegisterOracle'"
}

func (e RegisterOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	panic("implement me")
}

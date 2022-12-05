package event

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Event interface {
	GetEventQuery() string
	EventHandler(event ctypes.ResultEvent) error
}

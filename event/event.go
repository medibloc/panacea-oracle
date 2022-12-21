package event

import (
	"context"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Event interface {
	Name() string
	GetEventQuery() string
	EventHandler(context.Context, ctypes.ResultEvent) error
}

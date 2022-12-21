package event

import (
	"fmt"
	"strconv"

	"github.com/medibloc/panacea-oracle/panacea"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Event interface {
	Name() string
	GetEventQuery() string
	EventHandler(ctypes.ResultEvent) error
}

func GetQueryHeight(queryClient panacea.QueryClient, e ctypes.ResultEvent) (int64, error) {
	height, err := strconv.ParseInt(e.Events["tx.height"][0], 10, 64)
	if err != nil {
		log.Warn("failed to get height from event. Set to the height of the last block.")
		height, err = queryClient.GetLastBlockHeight()
		if err != nil {
			return 0, fmt.Errorf("failed to set query height. %v", err)
		}
	}

	return height, nil
}

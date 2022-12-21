package middleware

import (
	"net/http"

	"github.com/medibloc/panacea-oracle/panacea"
	log "github.com/sirupsen/logrus"
)

type queryHeightMiddleware struct {
	panaceaQueryClient panacea.QueryClient
}

func NewQueryMiddleWare(queryClient panacea.QueryClient) *queryHeightMiddleware {
	return &queryHeightMiddleware{
		panaceaQueryClient: queryClient,
	}
}

func (mw *queryHeightMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Retrieving the last block")
		height, err := mw.panaceaQueryClient.GetLastBlockHeight(r.Context())
		if err == nil {
			log.Debugf("Set the previous height of the last block height. LastHeight: %v, SetHeight: %v", height, height-1)
			r = r.WithContext(
				panacea.SetQueryBlockHeightToContext(r.Context(), height-1),
			)
		} else {
			log.Warnf("failed to get last block height. %v", err)
		}

		next.ServeHTTP(w, r)
	})
}

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
		height := mw.panaceaQueryClient.GetCachedLastBlockHeight()

		panacea.SetQueryBlockHeightToContext(r.Context(), height-1)

		next.ServeHTTP(w, r)
	})
}

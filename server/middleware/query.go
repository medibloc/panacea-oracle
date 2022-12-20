package middleware

import (
	"net/http"

	"github.com/medibloc/panacea-oracle/panacea"
	log "github.com/sirupsen/logrus"
)

type queryMiddleware struct {
	panaceaQueryClient panacea.QueryClient
}

func NewQueryMiddleWare(queryClient panacea.QueryClient) *queryMiddleware {
	return &queryMiddleware{
		panaceaQueryClient: queryClient,
	}
}

func (mw *queryMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Retrieving the last block")
		height, err := mw.panaceaQueryClient.GetLastBlockHeight(r.Context())
		if err == nil {
			log.Infof("Set the previous value of the last block heignt. LastHeight: %v, SetHeight: %v", height, height-1)
			r = r.WithContext(
				panacea.SetQueryBlockHeightToContext(r.Context(), height-1),
			)
		} else {
			log.Errorf("failed to get last block height. %v", err)
		}

		next.ServeHTTP(w, r)
	})
}

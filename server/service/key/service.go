package key

import (
	"net/http"

	"github.com/gorilla/mux"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
)

type combinedKeyService struct {
	serverservice.Reactor
}

func RegisterHandlers(svc serverservice.Reactor, router *mux.Router) {
	s := &combinedKeyService{svc}

	router.HandleFunc("/v0/data-deal/secret-key?deal-id={dealId}&data-hash={dataHash}}", s.GetSecretKey).Methods(http.MethodGet)
}

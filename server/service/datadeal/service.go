package datadeal

import (
	"net/http"

	"github.com/gorilla/mux"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
)

type dataDealService struct {
	serverservice.Reactor
}

func RegisterHandlers(svc serverservice.Reactor, router *mux.Router) {
	s := &dataDealService{svc}

	router.HandleFunc("/v0/data-deal/deals/{dealId}/data", s.ValidateData).Methods(http.MethodPost)
}
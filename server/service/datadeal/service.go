package datadeal

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/medibloc/panacea-oracle/service"
)

type dataDealService struct {
	service.Service
}

func RegisterHandlers(svc service.Service, router *mux.Router) {
	s := &dataDealService{svc}

	router.HandleFunc("/v0/data-deal/deals/{dealId}/data", s.ValidateData).Methods(http.MethodPost)
}

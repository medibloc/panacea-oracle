package status

import (
	"net/http"

	"github.com/gorilla/mux"
	serverservice "github.com/medibloc/panacea-oracle/server/service"
)

type statusService struct {
	serverservice.Service
}

func RegisterHandlers(svc serverservice.Service, router *mux.Router) {
	s := &statusService{svc}

	router.HandleFunc("/v0/status", s.GetStatus).Methods(http.MethodGet)
}
